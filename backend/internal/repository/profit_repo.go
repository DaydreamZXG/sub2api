package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// profitHourlyUpsertSQL sums usage logs created in [$1, $2) into hourly buckets per account, group and billing type.
// Buckets are UTC hours, so the same log always lands in the same bucket whatever the site's timezone is.
// paid_cost is the part of actual_cost that was real money: each user's charge times the share of their
// credits that came from paid orders (and, when the settings say so, from manual grants and redeem codes).
// Paid and gifted balance sit in one column, so they are taken to be spent in proportion.
const profitHourlyUpsertSQL = `
	WITH active AS (
		-- Only the users charged in these hours: this runs every minute.
		SELECT DISTINCT user_id FROM usage_logs WHERE created_at >= $1 AND created_at < $2
	), paid AS (
		SELECT user_id, SUM(GREATEST(amount - bonus_amount - refund_amount, 0)) AS amount
		FROM payment_orders
		WHERE order_type = 'balance' AND user_id IN (SELECT user_id FROM active)
			AND status IN ('COMPLETED', 'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING', 'PARTIALLY_REFUNDED', 'REFUNDED', 'REFUND_FAILED')
		GROUP BY user_id
	), granted AS (
		-- Net of what was taken back: an admin adding 600 and removing 600 again granted nothing.
		SELECT used_by AS user_id, GREATEST(SUM(value), 0) AS amount
		FROM redeem_codes
		WHERE status = 'used' AND used_by IN (SELECT user_id FROM active) AND (
			(type = 'admin_balance' AND COALESCE((SELECT value FROM settings WHERE key = 'profit_manual_is_paid'), '') = 'true')
			OR (type = 'balance' AND COALESCE((SELECT value FROM settings WHERE key = 'profit_redeem_is_paid'), '') = 'true'))
		GROUP BY used_by
	), share AS (
		SELECT u.id AS user_id,
			CASE WHEN u.total_recharged > 0
				THEN LEAST((COALESCE(p.amount, 0) + COALESCE(g.amount, 0)) / u.total_recharged, 1)
				ELSE 0 END AS paid_share
		FROM active a
		JOIN users u ON u.id = a.user_id
		LEFT JOIN paid p ON p.user_id = u.id
		LEFT JOIN granted g ON g.user_id = u.id
	)
	INSERT INTO profit_hourly (bucket_start, account_id, group_id, billing_type, requests, standard_cost, actual_cost, account_cost, paid_cost)
	SELECT
		date_trunc('hour', l.created_at AT TIME ZONE 'UTC') AT TIME ZONE 'UTC',
		l.account_id,
		COALESCE(l.group_id, 0),
		COALESCE(l.billing_type, 0),
		COUNT(*),
		COALESCE(SUM(l.total_cost), 0),
		COALESCE(SUM(l.actual_cost), 0),
		COALESCE(SUM(COALESCE(l.account_stats_cost, l.total_cost) * COALESCE(l.account_rate_multiplier, 1)), 0),
		COALESCE(SUM(l.actual_cost * COALESCE(s.paid_share, 0)), 0)
	FROM usage_logs l
	LEFT JOIN share s ON s.user_id = l.user_id
	WHERE l.created_at >= $1 AND l.created_at < $2
	GROUP BY 1, 2, 3, 4
	ON CONFLICT (bucket_start, account_id, group_id, billing_type) DO UPDATE SET
		requests = EXCLUDED.requests,
		standard_cost = EXCLUDED.standard_cost,
		actual_cost = EXCLUDED.actual_cost,
		account_cost = EXCLUDED.account_cost,
		paid_cost = EXCLUDED.paid_cost
	WHERE EXCLUDED.requests >= profit_hourly.requests`

// upsertProfitHourly refreshes the profit summary for whole hours. Saved figures only ever grow: an hour
// whose logs have since been purged, in full or in part, keeps what it had, because that money was charged.
func (r *dashboardAggregationRepository) upsertProfitHourly(ctx context.Context, hourStart, hourEnd time.Time) error {
	_, err := r.sql.ExecContext(ctx, profitHourlyUpsertSQL, hourStart, hourEnd)
	return err
}

type profitRepository struct {
	db *sql.DB
}

// NewProfitRepository creates the store for cost packs and the profit summary.
func NewProfitRepository(db *sql.DB) service.ProfitRepository {
	return &profitRepository{db: db}
}

const profitDateFormat = "2006-01-02"

func (r *profitRepository) ListPacks(ctx context.Context) ([]service.CostPack, error) {
	return r.listPacks(ctx, 0)
}

func (r *profitRepository) GetPack(ctx context.Context, id int64) (*service.CostPack, error) {
	packs, err := r.listPacks(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(packs) == 0 {
		return nil, service.ErrCostPackNotFound
	}
	return &packs[0], nil
}

// listPacks loads packs with their payments and accounts; id 0 loads all of them.
func (r *profitRepository) listPacks(ctx context.Context, id int64) ([]service.CostPack, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, note, cycle_days, cycle_amount, started_at, ended_at
		FROM cost_packs
		WHERE $1 = 0 OR id = $1
		ORDER BY (ended_at IS NOT NULL AND ended_at < CURRENT_DATE), started_at DESC, id DESC`, id)
	if err != nil {
		return nil, err
	}
	packs := make([]service.CostPack, 0)
	index := map[int64]int{}
	for rows.Next() {
		var (
			p       service.CostPack
			started time.Time
			ended   sql.NullTime
		)
		if err := rows.Scan(&p.ID, &p.Name, &p.Note, &p.CycleDays, &p.CycleAmount, &started, &ended); err != nil {
			_ = rows.Close()
			return nil, err
		}
		p.StartedAt = started.Format(profitDateFormat)
		if ended.Valid {
			v := ended.Time.Format(profitDateFormat)
			p.EndedAt = &v
		}
		p.Payments = []service.CostPackPayment{}
		p.Accounts = []service.CostPackAccount{}
		index[p.ID] = len(packs)
		packs = append(packs, p)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = r.db.QueryContext(ctx, `
		SELECT pack_id, id, amount, paid_at, note
		FROM cost_pack_payments
		WHERE $1 = 0 OR pack_id = $1
		ORDER BY paid_at, id`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			packID int64
			pay    service.CostPackPayment
			paidAt time.Time
		)
		if err := rows.Scan(&packID, &pay.ID, &pay.Amount, &paidAt, &pay.Note); err != nil {
			_ = rows.Close()
			return nil, err
		}
		pay.PaidAt = paidAt.Format(profitDateFormat)
		if i, ok := index[packID]; ok {
			packs[i].Payments = append(packs[i].Payments, pay)
		}
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = r.db.QueryContext(ctx, `
		SELECT cpa.pack_id, cpa.account_id, COALESCE(a.name, ''), COALESCE(a.platform, ''), (a.id IS NULL OR a.deleted_at IS NOT NULL)
		FROM cost_pack_accounts cpa
		LEFT JOIN accounts a ON a.id = cpa.account_id
		WHERE $1 = 0 OR cpa.pack_id = $1
		ORDER BY cpa.account_id`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			packID int64
			acc    service.CostPackAccount
		)
		if err := rows.Scan(&packID, &acc.ID, &acc.Name, &acc.Platform, &acc.Deleted); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if i, ok := index[packID]; ok {
			packs[i].Accounts = append(packs[i].Accounts, acc)
		}
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}
	return packs, nil
}

func closeRows(rows *sql.Rows) error {
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	return rows.Close()
}

func isProfitUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (r *profitRepository) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func nullableDate(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

// setPackAccounts replaces the accounts of a pack. An account that another pack holds is refused.
func setPackAccounts(ctx context.Context, tx *sql.Tx, packID int64, accountIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM cost_pack_accounts WHERE pack_id = $1`, packID); err != nil {
		return err
	}
	if len(accountIDs) == 0 {
		return nil
	}
	var known int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounts WHERE id = ANY($1)`, pq.Array(accountIDs)).Scan(&known); err != nil {
		return err
	}
	if known != len(accountIDs) {
		return service.ErrCostPackInvalid
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO cost_pack_accounts (pack_id, account_id)
		SELECT $1, unnest($2::bigint[])`, packID, pq.Array(accountIDs))
	if isProfitUniqueViolation(err) {
		return service.ErrCostPackAccountTaken
	}
	return err
}

func (r *profitRepository) CreatePack(ctx context.Context, in service.CostPackInput) (int64, error) {
	var id int64
	err := r.inTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO cost_packs (name, note, cycle_days, cycle_amount, started_at, ended_at)
			VALUES ($1, $2, $3, $4, $5::date, $6::date)
			RETURNING id`,
			in.Name, in.Note, in.CycleDays, in.CycleAmount, in.StartedAt, nullableDate(in.EndedAt)).Scan(&id); err != nil {
			return err
		}
		if in.CycleDays == 0 {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO cost_pack_payments (pack_id, amount, paid_at) VALUES ($1, $2, $3::date)`,
				id, in.Amount, in.StartedAt); err != nil {
				return err
			}
		}
		return setPackAccounts(ctx, tx, id, in.AccountIDs)
	})
	return id, err
}

func (r *profitRepository) UpdatePack(ctx context.Context, id int64, in service.CostPackInput) error {
	return r.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE cost_packs
			SET name = $2, note = $3, cycle_days = $4, cycle_amount = $5, started_at = $6::date, ended_at = $7::date, updated_at = NOW()
			WHERE id = $1`,
			id, in.Name, in.Note, in.CycleDays, in.CycleAmount, in.StartedAt, nullableDate(in.EndedAt))
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return service.ErrCostPackNotFound
		}
		return setPackAccounts(ctx, tx, id, in.AccountIDs)
	})
}

func (r *profitRepository) DeletePack(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM cost_packs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return service.ErrCostPackNotFound
	}
	return nil
}

func (r *profitRepository) AddPayment(ctx context.Context, packID int64, amount float64, paidAt, note string) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO cost_pack_payments (pack_id, amount, paid_at, note)
		SELECT id, $2, $3::date, $4 FROM cost_packs WHERE id = $1`, packID, amount, paidAt, note)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return service.ErrCostPackNotFound
	}
	return nil
}

func (r *profitRepository) DeletePayment(ctx context.Context, packID, paymentID int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM cost_pack_payments WHERE id = $2 AND pack_id = $1`, packID, paymentID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return service.ErrCostPackNotFound
	}
	return nil
}

func (r *profitRepository) Usage(ctx context.Context, from, to, tz string) ([]service.ProfitUsageRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH h AS (
			SELECT (bucket_start AT TIME ZONE $3)::date AS day, account_id, group_id, billing_type,
				requests, standard_cost, actual_cost, account_cost,
				-- Hours summed up before paid_cost existed count in full.
				COALESCE(paid_cost, actual_cost) AS paid_cost
			FROM profit_hourly
			WHERE bucket_start >= ($1::date::timestamp AT TIME ZONE $3)
				AND bucket_start < (($2::date + 1)::timestamp AT TIME ZONE $3)
		)
		SELECT
			to_char(h.day, 'YYYY-MM'),
			h.account_id,
			h.group_id,
			h.billing_type,
			COALESCE(CASE WHEN h.day >= p.started_at AND (p.ended_at IS NULL OR h.day <= p.ended_at) THEN p.id END, 0),
			SUM(h.requests),
			SUM(h.standard_cost),
			SUM(h.actual_cost),
			SUM(h.account_cost),
			SUM(h.paid_cost)
		FROM h
		LEFT JOIN cost_pack_accounts cpa ON cpa.account_id = h.account_id
		LEFT JOIN cost_packs p ON p.id = cpa.pack_id
		GROUP BY 1, 2, 3, 4, 5`, from, to, tz)
	if err != nil {
		return nil, err
	}
	out := make([]service.ProfitUsageRow, 0)
	for rows.Next() {
		var row service.ProfitUsageRow
		if err := rows.Scan(&row.Month, &row.AccountID, &row.GroupID, &row.BillingType, &row.PackID,
			&row.Requests, &row.StandardCost, &row.ActualCost, &row.AccountCost, &row.PaidCost); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out = append(out, row)
	}
	return out, closeRows(rows)
}

// Received sums what paid orders brought in and what was refunded in the period, in the currency paid.
func (r *profitRepository) Received(ctx context.Context, from, to, tz string) (float64, float64, error) {
	var received, refunded float64
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(pay_amount) FILTER (
				WHERE paid_at >= ($1::date::timestamp AT TIME ZONE $3) AND paid_at < (($2::date + 1)::timestamp AT TIME ZONE $3)), 0),
			COALESCE(SUM(refund_amount * pay_amount / NULLIF(amount, 0)) FILTER (
				WHERE refund_at >= ($1::date::timestamp AT TIME ZONE $3) AND refund_at < (($2::date + 1)::timestamp AT TIME ZONE $3)), 0)
		FROM payment_orders
		WHERE status IN ('COMPLETED', 'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING', 'PARTIALLY_REFUNDED', 'REFUNDED', 'REFUND_FAILED')`,
		from, to, tz).Scan(&received, &refunded)
	return received, refunded, err
}

func (r *profitRepository) PackUsage(ctx context.Context, tz string) (map[int64]service.ProfitPackUsage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,
			COALESCE(SUM(h.requests), 0),
			COALESCE(SUM(CASE WHEN h.billing_type = 1 THEN 0 ELSE COALESCE(h.paid_cost, h.actual_cost) END), 0),
			COALESCE(SUM(CASE WHEN h.billing_type = 1 THEN h.actual_cost ELSE 0 END), 0)
		FROM cost_packs p
		JOIN cost_pack_accounts cpa ON cpa.pack_id = p.id
		JOIN profit_hourly h ON h.account_id = cpa.account_id
			AND h.bucket_start >= (p.started_at::timestamp AT TIME ZONE $1)
			AND (p.ended_at IS NULL OR h.bucket_start < ((p.ended_at + 1)::timestamp AT TIME ZONE $1))
		GROUP BY p.id`, tz)
	if err != nil {
		return nil, err
	}
	out := map[int64]service.ProfitPackUsage{}
	for rows.Next() {
		var (
			id int64
			u  service.ProfitPackUsage
		)
		if err := rows.Scan(&id, &u.Requests, &u.Income, &u.PlanUsage); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out[id] = u
	}
	return out, closeRows(rows)
}

func (r *profitRepository) AccountInfos(ctx context.Context, ids []int64) (map[int64]service.ProfitAccountInfo, error) {
	out := map[int64]service.ProfitAccountInfo{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, platform, rate_multiplier, deleted_at IS NOT NULL
		FROM accounts WHERE id = ANY($1)`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			id   int64
			info service.ProfitAccountInfo
		)
		if err := rows.Scan(&id, &info.Name, &info.Platform, &info.RateMultiplier, &info.Deleted); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out[id] = info
	}
	return out, closeRows(rows)
}

func (r *profitRepository) GroupNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	out := map[int64]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM groups WHERE id = ANY($1)`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			id   int64
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out[id] = name
	}
	return out, closeRows(rows)
}

// RebuildSummary folds every usage log still kept into the summary.
func (r *profitRepository) RebuildSummary(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, profitHourlyUpsertSQL, time.Unix(0, 0).UTC(), time.Now().UTC().Add(time.Hour))
	return err
}
