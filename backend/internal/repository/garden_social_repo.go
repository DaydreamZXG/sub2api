package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type gardenSocialRepository struct {
	db *sql.DB
}

// NewGardenSocialRepository creates the store for what neighbours do in each other's gardens.
func NewGardenSocialRepository(db *sql.DB) service.GardenSocialRepository {
	return &gardenSocialRepository{db: db}
}

// 固定邻居最多取这么多个邀请来的人，免得邀请很多人的用户列表过长。
const gardenPinnedInvitees = 20

func (r *gardenSocialRepository) Neighbours(ctx context.Context, viewerID int64, perDay int) (neighbours []service.GardenNeighbour, err error) {
	// 随机邻居按 (对方, 自己, 日期) 的哈希排序取前几名：同一天内稳定，第二天换一批。
	const query = `
		WITH pinned AS (
			SELECT inviter_id AS id FROM user_affiliates WHERE user_id = $1 AND inviter_id IS NOT NULL
			UNION
			SELECT user_id FROM (
				SELECT user_id FROM user_affiliates WHERE inviter_id = $1 ORDER BY created_at DESC LIMIT $3
			) invited
		), drawn AS (
			SELECT c.user_id AS id
			FROM garden_crops c
			JOIN users owner ON owner.id = c.user_id AND owner.status = 'active' AND owner.deleted_at IS NULL
			WHERE c.status = 'growing' AND c.user_id <> $1 AND c.user_id NOT IN (SELECT id FROM pinned)
			GROUP BY c.user_id
			ORDER BY md5(c.user_id::text || ':' || $1::text || ':' || CURRENT_DATE::text)
			LIMIT $2
		)
		SELECT u.id, u.username, u.email, u.id IN (SELECT id FROM pinned) AS is_pinned,
			(SELECT COUNT(*) FROM garden_crops c WHERE c.user_id = u.id AND c.status = 'growing')
		FROM users u
		WHERE u.id IN (SELECT id FROM pinned UNION SELECT id FROM drawn)
			AND u.id <> $1 AND u.status = 'active' AND u.deleted_at IS NULL
		ORDER BY is_pinned DESC, md5(u.id::text || ':' || $1::text || ':' || CURRENT_DATE::text)
	`
	rows, err := r.db.QueryContext(ctx, query, viewerID, perDay, gardenPinnedInvitees)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()
	neighbours = []service.GardenNeighbour{}
	for rows.Next() {
		var n service.GardenNeighbour
		if err = rows.Scan(&n.ID, &n.Username, &n.Email, &n.Pinned, &n.Crops); err != nil {
			return nil, err
		}
		neighbours = append(neighbours, n)
	}
	return neighbours, rows.Err()
}

func (r *gardenSocialRepository) NeighbourCrops(ctx context.Context, ownerID, viewerID int64) (crops []service.GardenNeighbourCropState, err error) {
	const query = `
		SELECT c.id, c.amount, c.stolen_amount, c.planted_at, c.ripe_at,
			(SELECT COUNT(*) FROM garden_actions a WHERE a.crop_id = c.id AND a.kind = 'water'),
			EXISTS (SELECT 1 FROM garden_actions a
				WHERE a.crop_id = c.id AND a.actor_id = $2 AND a.kind = 'water' AND a.created_at >= CURRENT_DATE),
			COALESCE((SELECT a.kind FROM garden_actions a
				WHERE a.crop_id = c.id AND a.actor_id = $2 AND a.kind IN ('steal', 'caught') LIMIT 1), '')
		FROM garden_crops c
		WHERE c.user_id = $1 AND c.status = 'growing'
		ORDER BY c.ripe_at, c.id
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID, viewerID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()
	for rows.Next() {
		var c service.GardenNeighbourCropState
		if err = rows.Scan(&c.ID, &c.Amount, &c.StolenAmount, &c.PlantedAt, &c.RipeAt, &c.Waters, &c.WateredToday, &c.Tried); err != nil {
			return nil, err
		}
		crops = append(crops, c)
	}
	return crops, rows.Err()
}

func (r *gardenSocialRepository) CropOwner(ctx context.Context, cropID int64) (ownerID int64, totalRecharged float64, ok bool, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT c.user_id, u.total_recharged
		FROM garden_crops c JOIN users u ON u.id = c.user_id
		WHERE c.id = $1 AND c.status = 'growing'`, cropID).Scan(&ownerID, &totalRecharged)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return ownerID, totalRecharged, err == nil, err
}

// lockActor serialises one user's garden actions: their dog row is created if missing and locked for the transaction.
func lockActor(ctx context.Context, tx *sql.Tx, actorID int64) error {
	// feed_seen_at 从很早的时间起算：只有读动态才往前推，建这一行不能把还没看的动态算成看过。
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO garden_dogs (user_id, feed_seen_at) VALUES ($1, 'epoch') ON CONFLICT (user_id) DO NOTHING`, actorID); err != nil {
		return err
	}
	var locked int64
	return tx.QueryRowContext(ctx, `SELECT user_id FROM garden_dogs WHERE user_id = $1 FOR UPDATE`, actorID).Scan(&locked)
}

func (r *gardenSocialRepository) Steal(ctx context.Context, actorID, cropID int64, rules service.GardenStealRules) (result service.GardenStealResult, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = lockActor(ctx, tx, actorID); err != nil {
		return result, err
	}

	var (
		ownerID        int64
		amount, stolen float64
		open           bool
	)
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, amount, stolen_amount, ripe_at + make_interval(secs => $2) <= NOW()
		FROM garden_crops WHERE id = $1 AND status = 'growing' FOR UPDATE`,
		cropID, rules.ProtectFor.Seconds()).Scan(&ownerID, &amount, &stolen, &open)
	if errors.Is(err, sql.ErrNoRows) {
		return result, service.ErrGardenCropNotFound
	}
	if err != nil {
		return result, err
	}
	if ownerID == actorID {
		return result, service.ErrGardenOwnCrop
	}
	if !open {
		return result, service.ErrGardenCropProtected
	}

	var (
		hasCrop, tried bool
		triesToday     int
	)
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM garden_crops WHERE user_id = $1 AND status = 'growing'),
			(SELECT COUNT(*) FROM garden_actions
				WHERE actor_id = $1 AND kind IN ('steal', 'caught') AND created_at >= CURRENT_DATE),
			EXISTS (SELECT 1 FROM garden_actions WHERE actor_id = $1 AND crop_id = $2 AND kind IN ('steal', 'caught'))`,
		actorID, cropID).Scan(&hasCrop, &triesToday, &tried)
	if err != nil {
		return result, err
	}
	switch {
	case !hasCrop:
		return result, service.ErrGardenNeedCrop
	case tried:
		return result, service.ErrGardenAlreadyTried
	case triesToday >= rules.DailyLimit:
		return result, service.ErrGardenStealLimit
	}

	// 余额最小单位是 1e-8，向下取整，偷走的总和不会超过上限。
	take := math.Floor(math.Min(amount*rules.StealShare, amount*rules.MaxStolenShare-stolen)*1e8) / 1e8
	if take <= 0 {
		return result, service.ErrGardenPickedClean
	}

	if rules.DogCatches {
		err = tx.QueryRowContext(ctx, `
			SELECT COALESCE((SELECT fed_on = CURRENT_DATE FROM garden_dogs WHERE user_id = $1), FALSE)`,
			ownerID).Scan(&result.Caught)
		if err != nil {
			return result, err
		}
	}
	kind := service.GardenActionCaught
	if !result.Caught {
		kind = service.GardenActionSteal
		result.Amount = take
		if _, err = tx.ExecContext(ctx, `UPDATE garden_crops SET stolen_amount = stolen_amount + $2 WHERE id = $1`, cropID, take); err != nil {
			return result, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $2, updated_at = NOW() WHERE id = $1`, actorID, take); err != nil {
			return result, err
		}
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO garden_actions (crop_id, owner_id, actor_id, kind, amount) VALUES ($1, $2, $3, $4, $5)`,
		cropID, ownerID, actorID, kind, result.Amount); err != nil {
		return result, err
	}
	return result, tx.Commit()
}

func (r *gardenSocialRepository) Water(ctx context.Context, actorID, cropID int64, sooner time.Duration, maxWaters int) (ripeAt time.Time, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ripeAt, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		ownerID int64
		growing bool
	)
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, ripe_at > NOW() FROM garden_crops WHERE id = $1 AND status = 'growing' FOR UPDATE`,
		cropID).Scan(&ownerID, &growing)
	if errors.Is(err, sql.ErrNoRows) {
		return ripeAt, service.ErrGardenCropNotFound
	}
	if err != nil {
		return ripeAt, err
	}
	if ownerID == actorID {
		return ripeAt, service.ErrGardenOwnCrop
	}
	if !growing {
		return ripeAt, service.ErrGardenCropRipe
	}

	var waters, mineToday int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE actor_id = $2 AND created_at >= CURRENT_DATE)
		FROM garden_actions WHERE crop_id = $1 AND kind = 'water'`, cropID, actorID).Scan(&waters, &mineToday)
	if err != nil {
		return ripeAt, err
	}
	if mineToday > 0 {
		return ripeAt, service.ErrGardenAlreadyWatered
	}
	if waters >= maxWaters {
		return ripeAt, service.ErrGardenWaterFull
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE garden_crops SET ripe_at = GREATEST(ripe_at - make_interval(secs => $2), NOW())
		WHERE id = $1 RETURNING ripe_at`, cropID, sooner.Seconds()).Scan(&ripeAt)
	if err != nil {
		return ripeAt, err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO garden_actions (crop_id, owner_id, actor_id, kind) VALUES ($1, $2, $3, 'water')`,
		cropID, ownerID, actorID); err != nil {
		return ripeAt, err
	}
	return ripeAt, tx.Commit()
}

func (r *gardenSocialRepository) FeedDog(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO garden_dogs (user_id, fed_on, feed_seen_at) VALUES ($1, CURRENT_DATE, 'epoch')
		ON CONFLICT (user_id) DO UPDATE SET fed_on = CURRENT_DATE`, userID)
	return err
}

func (r *gardenSocialRepository) Social(ctx context.Context, userID int64) (social service.GardenSocial, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE((SELECT fed_on = CURRENT_DATE FROM garden_dogs WHERE user_id = $1), FALSE),
			(SELECT COUNT(*) FROM garden_actions
				WHERE actor_id = $1 AND kind IN ('steal', 'caught') AND created_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM garden_actions
				WHERE owner_id = $1 AND created_at > COALESCE(
					(SELECT feed_seen_at FROM garden_dogs WHERE user_id = $1), 'epoch'::timestamptz)),
			COALESCE((SELECT total_recharged FROM users WHERE id = $1), 0)`,
		userID).Scan(&social.DogFedToday, &social.StealsToday, &social.Unread, &social.TotalRecharged)
	return social, err
}

func (r *gardenSocialRepository) Feed(ctx context.Context, userID int64, limit int) (items []service.GardenFeedItem, err error) {
	const query = `
		SELECT a.kind, a.amount, a.created_at, a.actor_id = $1, u.username, u.email
		FROM garden_actions a
		JOIN users u ON u.id = CASE WHEN a.actor_id = $1 THEN a.owner_id ELSE a.actor_id END
		WHERE a.owner_id = $1 OR a.actor_id = $1
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()
	items = []service.GardenFeedItem{}
	for rows.Next() {
		var item service.GardenFeedItem
		if err = rows.Scan(&item.Kind, &item.Amount, &item.CreatedAt, &item.Mine, &item.Username, &item.Email); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *gardenSocialRepository) MarkFeedSeen(ctx context.Context, userID int64, until time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO garden_dogs (user_id, feed_seen_at) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET feed_seen_at = GREATEST(garden_dogs.feed_seen_at, EXCLUDED.feed_seen_at)`,
		userID, until)
	return err
}
