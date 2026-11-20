package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const agentWatermarkKey = "agent_share_watermark"

type agentRepository struct {
	client *dbent.Client
	db     *sql.DB
}

// NewAgentRepository creates the store for agents, their shares and their payout requests.
func NewAgentRepository(client *dbent.Client, db *sql.DB) service.AgentRepository {
	return &agentRepository{client: client, db: db}
}

func (r *agentRepository) SetAgent(ctx context.Context, userID int64, isAgent bool, sharePercent float64, tiers []service.AgentTier) error {
	var ladder any
	if isAgent && len(tiers) > 0 {
		encoded, err := json.Marshal(tiers)
		if err != nil {
			return err
		}
		ladder = string(encoded)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE user_affiliates
		SET agent_since = CASE WHEN NOT $2 THEN NULL WHEN is_agent THEN agent_since ELSE NOW() END,
			agent_share_percent = CASE WHEN $2 THEN $3::numeric ELSE NULL END,
			agent_tiers = $4::jsonb,
			is_agent = $2,
			updated_at = NOW()
		WHERE user_id = $1`, userID, isAgent, sharePercent, ladder)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return service.ErrUserNotFound
	}
	return nil
}

// agentTiersFromJSON reads a stored ladder; an agent without one has a flat share.
func agentTiersFromJSON(raw sql.NullString) ([]service.AgentTier, error) {
	tiers := []service.AgentTier{}
	if !raw.Valid || raw.String == "" {
		return tiers, nil
	}
	if err := json.Unmarshal([]byte(raw.String), &tiers); err != nil {
		return nil, err
	}
	return tiers, nil
}

func (r *agentRepository) ListAgents(ctx context.Context, monthStart time.Time) ([]service.AgentEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ua.user_id, COALESCE(u.email, ''), COALESCE(ua.agent_share_percent, 0), ua.agent_since,
			(SELECT COUNT(*) FROM user_affiliates i WHERE i.inviter_id = ua.user_id),
			COALESCE((SELECT SUM(l.amount) FROM user_affiliate_ledger l WHERE l.user_id = ua.user_id AND l.action = 'agent_share'), 0),
			ua.aff_quota, ua.agent_tiers::text,
			COALESCE((SELECT SUM(l.base_amount) FROM user_affiliate_ledger l WHERE l.user_id = ua.user_id AND l.action = 'agent_share' AND l.created_at >= $1), 0)
		FROM user_affiliates ua
		LEFT JOIN users u ON u.id = ua.user_id
		WHERE ua.is_agent
		ORDER BY ua.agent_since DESC NULLS LAST, ua.user_id`, monthStart)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.AgentEntry, 0)
	for rows.Next() {
		var e service.AgentEntry
		var since sql.NullTime
		var ladder sql.NullString
		if err := rows.Scan(&e.UserID, &e.Email, &e.SharePercent, &since, &e.Invitees, &e.Shared, &e.Available, &ladder, &e.MonthSpent); err != nil {
			return nil, err
		}
		if e.Tiers, err = agentTiersFromJSON(ladder); err != nil {
			return nil, err
		}
		if since.Valid {
			e.Since = &since.Time
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// agentAccrueSQL credits, for the usage in ($1, $2], each agent their share of the real money their
// users were charged, one ledger row per agent and user. $3 limits it to one user, 0 means all.
//
// An agent with tiers earns each part of the spending at the share of the band it falls in: the bands
// are counted over what was credited since the month began ($4), so nothing credited earlier changes.
const agentAccrueSQL = `
	WITH active AS (
		SELECT DISTINCT l.user_id
		FROM usage_logs l
		JOIN user_affiliates ua ON ua.user_id = l.user_id
		JOIN user_affiliates ag ON ag.user_id = ua.inviter_id AND ag.is_agent
		WHERE l.created_at > $1 AND l.created_at <= $2
			AND ($3::bigint = 0 OR l.user_id = $3)
	), ` + profitPaidShareCTEs + `, due AS (
		SELECT ag.user_id AS agent_id, l.user_id AS invitee_id,
			SUM(l.actual_cost * COALESCE(s.paid_share, 0)) AS base
		FROM usage_logs l
		JOIN user_affiliates ua ON ua.user_id = l.user_id
		JOIN user_affiliates ag ON ag.user_id = ua.inviter_id AND ag.is_agent AND ag.agent_share_percent > 0
		LEFT JOIN share s ON s.user_id = l.user_id
		WHERE l.created_at > $1 AND l.created_at <= $2
			AND COALESCE(l.billing_type, 0) = 0
			AND ($3::bigint = 0 OR l.user_id = $3)
			AND l.created_at >= ag.agent_since
			AND (ua.agent_bound_at IS NULL OR l.created_at > ua.agent_bound_at)
		GROUP BY ag.user_id, l.user_id
	), run AS (
		SELECT d.agent_id, SUM(d.base) AS base,
			COALESCE((SELECT SUM(m.base_amount) FROM user_affiliate_ledger m
				WHERE m.user_id = d.agent_id AND m.action = 'agent_share' AND m.created_at >= $4), 0) AS before
		FROM due d
		GROUP BY d.agent_id
	), band AS (
		SELECT agent_id, lo, pct, LEAD(lo) OVER (PARTITION BY agent_id ORDER BY lo) AS hi
		FROM (
			SELECT ag.user_id AS agent_id, 0::numeric AS lo, ag.agent_share_percent AS pct
			FROM user_affiliates ag JOIN run r ON r.agent_id = ag.user_id
			UNION ALL
			SELECT ag.user_id, (t->>'from')::numeric, (t->>'percent')::numeric
			FROM user_affiliates ag JOIN run r ON r.agent_id = ag.user_id
			CROSS JOIN LATERAL jsonb_array_elements(
				CASE WHEN jsonb_typeof(ag.agent_tiers) = 'array' THEN ag.agent_tiers ELSE '[]'::jsonb END) t
		) b
	), earned AS (
		SELECT r.agent_id, r.base,
			SUM(GREATEST(LEAST(r.before + r.base, COALESCE(b.hi, r.before + r.base)) - GREATEST(r.before, b.lo), 0) * b.pct / 100) AS amount
		FROM run r JOIN band b ON b.agent_id = r.agent_id
		GROUP BY r.agent_id, r.base
	), credited AS (
		INSERT INTO user_affiliate_ledger (user_id, action, amount, base_amount, source_user_id, created_at, updated_at)
		SELECT d.agent_id, 'agent_share', ROUND(e.amount * d.base / e.base, 8), d.base, d.invitee_id, $5, $5
		FROM due d JOIN earned e ON e.agent_id = d.agent_id
		WHERE e.base > 0 AND ROUND(e.amount * d.base / e.base, 8) > 0
		RETURNING user_id, amount
	)
	UPDATE user_affiliates ua
	SET aff_quota = ua.aff_quota + t.amount,
		aff_history_quota = ua.aff_history_quota + t.amount,
		updated_at = NOW()
	FROM (SELECT user_id, SUM(amount) AS amount FROM credited GROUP BY user_id) t
	WHERE ua.user_id = t.user_id`

// lockAgentWatermark returns the moment up to which usage was shared, and holds its row until the
// transaction ends: one accrual, or one move of a user between agents, at a time. The first call only
// sets the mark to firstMark: nothing from before the feature is shared.
func lockAgentWatermark(ctx context.Context, tx *sql.Tx, firstMark time.Time) (time.Time, error) {
	mark := firstMark.UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING`, agentWatermarkKey, mark); err != nil {
		return time.Time{}, err
	}
	var stored string
	if err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1 FOR UPDATE`, agentWatermarkKey).Scan(&stored); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, stored)
}

// lockAgentMark takes the mark's lock and returns the mark with the present moment. The moment is read
// from the database after the lock is held, so of two transactions the one that waited always gets the
// later moment, whichever server it ran on, and a wait across midnight on the last day of a month ends
// in the new month.
func lockAgentMark(ctx context.Context, tx *sql.Tx, firstMark time.Time) (from, now time.Time, err error) {
	if from, err = lockAgentWatermark(ctx, tx, firstMark); err != nil {
		return from, now, err
	}
	err = tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now)
	return from, now, err
}

// lockAgentMove is lockAgentMark for a change of a user's agent: now is the moment of the change.
func lockAgentMove(ctx context.Context, tx *sql.Tx) (from, now time.Time, err error) {
	return lockAgentMark(ctx, tx, time.Now())
}

// settleAgentUser credits the user's agent for what the user spent in (from, now], ahead of the regular
// run, because the user is about to leave that agent. from is the locked mark, which stays where it is:
// the regular run skips this usage, as by then the user has another agent from now on, or none.
func settleAgentUser(ctx context.Context, tx *sql.Tx, userID int64, from, now time.Time) error {
	if !now.After(from) {
		return nil
	}
	_, err := creditAgents(ctx, tx, from, now, userID, now)
	return err
}

// creditAgents shares the usage in (from, upTo], of one user or of all with userID 0. at is the moment
// the credits are dated, read after the mark's lock was taken: the month whose spending decides the
// tier is the month of at, so a credit always counts in the month whose tier it was worked out with.
func creditAgents(ctx context.Context, tx *sql.Tx, from, upTo time.Time, userID int64, at time.Time) (sql.Result, error) {
	return tx.ExecContext(ctx, agentAccrueSQL, from, upTo, userID, service.AgentMonthStart(at), at)
}

func (r *agentRepository) AccrueShares(ctx context.Context, upTo time.Time) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	from, at, err := lockAgentMark(ctx, tx, upTo)
	if err != nil {
		return 0, err
	}
	if !upTo.After(from) {
		return 0, tx.Commit()
	}
	res, err := creditAgents(ctx, tx, from, upTo, 0, at)
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE settings SET value = $2, updated_at = NOW() WHERE key = $1`, agentWatermarkKey, upTo.UTC().Format(time.RFC3339Nano)); err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), tx.Commit()
}

func (r *agentRepository) Overview(ctx context.Context, userID int64, monthStart time.Time) (*service.AgentOverview, error) {
	out := &service.AgentOverview{Invitees: []service.AgentInvitee{}, Payouts: []service.AgentPayout{}}
	var since sql.NullTime
	var ladder sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(ua.agent_share_percent, 0), ua.agent_since, ua.aff_quota, ua.agent_tiers::text,
			COALESCE(SUM(l.amount), 0), COALESCE(SUM(l.base_amount), 0),
			COALESCE(SUM(l.amount) FILTER (WHERE l.created_at >= $2), 0),
			COALESCE(SUM(l.base_amount) FILTER (WHERE l.created_at >= $2), 0)
		FROM user_affiliates ua
		LEFT JOIN user_affiliate_ledger l ON l.user_id = ua.user_id AND l.action = 'agent_share'
		WHERE ua.user_id = $1 AND ua.is_agent
		GROUP BY ua.user_id`, userID, monthStart).
		Scan(&out.SharePercent, &since, &out.Available, &ladder, &out.Shared, &out.Spent, &out.MonthShared, &out.MonthSpent)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if since.Valid {
		out.Since = &since.Time
	}
	if out.Tiers, err = agentTiersFromJSON(ladder); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT i.user_id, COALESCE(u.email, ''), i.created_at, COALESCE(SUM(l.base_amount), 0), COALESCE(SUM(l.amount), 0)
		FROM user_affiliates i
		LEFT JOIN users u ON u.id = i.user_id
		LEFT JOIN user_affiliate_ledger l ON l.user_id = $1 AND l.action = 'agent_share' AND l.source_user_id = i.user_id
		WHERE i.inviter_id = $1
		GROUP BY i.user_id, u.email, i.created_at
		ORDER BY 5 DESC, i.created_at DESC
		LIMIT 200`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var i service.AgentInvitee
		if err := rows.Scan(&i.UserID, &i.Email, &i.JoinedAt, &i.Spent, &i.Share); err != nil {
			return nil, err
		}
		out.Invitees = append(out.Invitees, i)
	}
	return out, rows.Err()
}

func (r *agentRepository) CreatePayout(ctx context.Context, userID int64, amount float64, payeeNote string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO agent_payout_requests (user_id, amount, payee_note)
		SELECT ua.user_id, $2::numeric, $3 FROM user_affiliates ua
		WHERE ua.user_id = $1 AND ua.is_agent AND ua.aff_quota >= $2::numeric
		RETURNING id`, userID, amount, payeeNote).Scan(&id)
	if err == nil {
		return id, nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return 0, service.ErrAgentPayoutPending
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	var isAgent bool
	if qErr := r.db.QueryRowContext(ctx, `SELECT is_agent FROM user_affiliates WHERE user_id = $1`, userID).Scan(&isAgent); qErr != nil && !errors.Is(qErr, sql.ErrNoRows) {
		return 0, qErr
	}
	if !isAgent {
		return 0, service.ErrAgentNotAgent
	}
	return 0, service.ErrAgentPayoutAmount
}

const agentPayoutColumns = `p.id, p.user_id, COALESCE(u.email, ''), p.amount, p.payee_note, p.status, p.admin_note, p.created_at, p.handled_at
	FROM agent_payout_requests p LEFT JOIN users u ON u.id = p.user_id`

func scanAgentPayout(scan func(...any) error) (service.AgentPayout, error) {
	var p service.AgentPayout
	var handled sql.NullTime
	err := scan(&p.ID, &p.UserID, &p.Email, &p.Amount, &p.PayeeNote, &p.Status, &p.AdminNote, &p.CreatedAt, &handled)
	if handled.Valid {
		p.HandledAt = &handled.Time
	}
	return p, err
}

func (r *agentRepository) ListPayouts(ctx context.Context, userID int64, status string, limit int) ([]service.AgentPayout, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+agentPayoutColumns+`
		WHERE ($1 = 0 OR p.user_id = $1) AND ($2 = '' OR p.status = $2)
		ORDER BY (p.status = 'pending') DESC, p.created_at DESC
		LIMIT $3`, userID, status, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.AgentPayout, 0)
	for rows.Next() {
		p, err := scanAgentPayout(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *agentRepository) GetPayout(ctx context.Context, id int64) (*service.AgentPayout, error) {
	p, err := scanAgentPayout(r.db.QueryRowContext(ctx, `SELECT `+agentPayoutColumns+` WHERE p.id = $1`, id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAgentPayoutNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *agentRepository) FinishPayout(ctx context.Context, id int64, status, adminNote string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE agent_payout_requests SET status = $2, admin_note = $3, handled_at = NOW()
		WHERE id = $1 AND status = 'pending'`, id, status, adminNote)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *agentRepository) PayPayout(ctx context.Context, id int64, adminNote string, withdraw func(txCtx context.Context) error) (bool, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin payout transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// The affiliate store joins the transaction it finds in the context.
	txCtx := dbent.NewTxContext(ctx, tx)
	res, err := tx.Client().ExecContext(txCtx, `
		UPDATE agent_payout_requests SET status = 'paid', admin_note = $2, handled_at = NOW()
		WHERE id = $1 AND status = 'pending'`, id, adminNote)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}
	if err := withdraw(txCtx); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit payout transaction: %w", err)
	}
	return true, nil
}

func (r *agentRepository) AgentUsers(ctx context.Context, agentID int64) ([]service.AgentInvitee, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.user_id, COALESCE(u.email, ''), COALESCE(i.agent_bound_at, i.created_at), i.agent_bound_at IS NOT NULL,
			COALESCE(SUM(l.base_amount), 0), COALESCE(SUM(l.amount), 0)
		FROM user_affiliates i
		LEFT JOIN users u ON u.id = i.user_id
		LEFT JOIN user_affiliate_ledger l ON l.user_id = $1 AND l.action = 'agent_share' AND l.source_user_id = i.user_id
		WHERE i.inviter_id = $1
		GROUP BY i.user_id, u.email, i.created_at, i.agent_bound_at
		ORDER BY 6 DESC, 3 DESC
		LIMIT 500`, agentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	users := []service.AgentInvitee{}
	for rows.Next() {
		var u service.AgentInvitee
		if err := rows.Scan(&u.UserID, &u.Email, &u.JoinedAt, &u.Attached, &u.Spent, &u.Share); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// AttachUser expects the user's affiliate row to exist already. The agent the user leaves is credited
// for the user's spending up to this moment; the new one earns on what comes after.
func (r *agentRepository) AttachUser(ctx context.Context, agentID, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Held first by every move and every accrual, so two moves cannot wait on each other's rows.
	from, now, err := lockAgentMove(ctx, tx)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id, is_agent, inviter_id FROM user_affiliates
		WHERE user_id IN ($1, $2) ORDER BY user_id FOR UPDATE`, agentID, userID)
	if err != nil {
		return err
	}
	var agentFound, userFound, isAgent bool
	var agentInviter, userInviter sql.NullInt64
	for rows.Next() {
		var id int64
		var agent bool
		var inviter sql.NullInt64
		if err := rows.Scan(&id, &agent, &inviter); err != nil {
			_ = rows.Close()
			return err
		}
		if id == agentID {
			agentFound, isAgent, agentInviter = true, agent, inviter
		} else {
			userFound, userInviter = true, inviter
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !agentFound || !isAgent {
		return service.ErrAgentNotAgent
	}
	// An agent who was brought in by this user cannot have them as their own user in turn.
	if !userFound || (agentInviter.Valid && agentInviter.Int64 == userID) {
		return service.ErrAgentUserInvalid
	}
	if userInviter.Valid && userInviter.Int64 == agentID {
		return nil
	}
	if userInviter.Valid {
		if err := settleAgentUser(ctx, tx, userID, from, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE user_affiliates SET aff_count = GREATEST(aff_count - 1, 0), updated_at = NOW() WHERE user_id = $1`, userInviter.Int64); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_affiliates SET inviter_id = $1, agent_bound_at = $3, updated_at = NOW() WHERE user_id = $2`, agentID, userID, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_affiliates SET aff_count = aff_count + 1, updated_at = NOW() WHERE user_id = $1`, agentID); err != nil {
		return err
	}
	return tx.Commit()
}

// DetachUser credits the agent for the user's spending up to this moment, then lets the user go.
func (r *agentRepository) DetachUser(ctx context.Context, agentID, userID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	from, now, err := lockAgentMove(ctx, tx)
	if err != nil {
		return false, err
	}
	var under bool
	if err := tx.QueryRowContext(ctx, `SELECT true FROM user_affiliates WHERE user_id = $2 AND inviter_id = $1 FOR UPDATE`, agentID, userID).Scan(&under); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if err := settleAgentUser(ctx, tx, userID, from, now); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE user_affiliates SET inviter_id = NULL, agent_bound_at = NULL, updated_at = NOW()
		WHERE user_id = $2 AND inviter_id = $1`, agentID, userID); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_affiliates SET aff_count = GREATEST(aff_count - 1, 0), updated_at = NOW() WHERE user_id = $1`, agentID); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
