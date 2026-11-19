package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type gardenRepository struct {
	client *dbent.Client
	db     *sql.DB
}

// NewGardenRepository creates the garden store.
func NewGardenRepository(client *dbent.Client, db *sql.DB) service.GardenRepository {
	return &gardenRepository{client: client, db: db}
}

const gardenCropColumns = `id, user_id, COALESCE(order_id, 0), amount, stolen_amount, status, planted_at, ripe_at, harvested_at, harvested_amount, source`

func scanGardenCrop(row interface{ Scan(dest ...any) error }) (*service.GardenCrop, error) {
	var (
		crop        service.GardenCrop
		harvestedAt sql.NullTime
	)
	if err := row.Scan(&crop.ID, &crop.UserID, &crop.OrderID, &crop.Amount, &crop.StolenAmount, &crop.Status,
		&crop.PlantedAt, &crop.RipeAt, &harvestedAt, &crop.HarvestedAmount, &crop.Source); err != nil {
		return nil, err
	}
	if harvestedAt.Valid {
		crop.HarvestedAt = &harvestedAt.Time
	}
	return &crop, nil
}

// movedAmount runs a statement that returns the single amount it moved, or no row when nothing matched.
func (r *gardenRepository) movedAmount(ctx context.Context, query string, args ...any) (float64, bool, error) {
	var amount float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&amount)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return amount, true, nil
}

func (r *gardenRepository) Plant(ctx context.Context, userID, orderID int64, amount float64, ripenAfter time.Duration) (bool, error) {
	// 种菜和扣余额在同一条语句里：订单已有菜时什么都不动，重复履约不会重复扣。
	const query = `
		WITH planted AS (
			INSERT INTO garden_crops (user_id, order_id, source, amount, ripe_at)
			VALUES ($1, $2, 'order', $3, NOW() + make_interval(secs => $4))
			ON CONFLICT (order_id) WHERE order_id IS NOT NULL DO NOTHING
			RETURNING user_id, amount
		)
		UPDATE users AS u
		SET balance = u.balance - planted.amount,
			total_recharged = GREATEST(u.total_recharged - planted.amount, 0),
			updated_at = NOW()
		FROM planted
		WHERE u.id = planted.user_id
		RETURNING planted.amount
	`
	_, ok, err := r.movedAmount(ctx, query, userID, orderID, amount, ripenAfter.Seconds())
	return ok, err
}

func (r *gardenRepository) Uproot(ctx context.Context, orderID int64) (float64, error) {
	const query = `
		WITH removed AS (
			UPDATE garden_crops
			SET status = 'removed'
			WHERE order_id = $1 AND status = 'growing'
			RETURNING user_id, amount - stolen_amount AS left_amount
		)
		UPDATE users AS u
		SET balance = u.balance + removed.left_amount,
			total_recharged = u.total_recharged + removed.left_amount,
			updated_at = NOW()
		FROM removed
		WHERE u.id = removed.user_id
		RETURNING removed.left_amount
	`
	amount, _, err := r.movedAmount(ctx, query, orderID)
	return amount, err
}

func (r *gardenRepository) GrowingForOrder(ctx context.Context, orderID int64) (float64, error) {
	amount, _, err := r.movedAmount(ctx, `
		SELECT amount - stolen_amount FROM garden_crops WHERE order_id = $1 AND status = 'growing'`, orderID)
	return amount, err
}

// Replant 在 ctx 带着事务时并入该事务（随它提交或回滚），否则自己开一个。
func (r *gardenRepository) Replant(ctx context.Context, orderID int64) (replanted float64, err error) {
	if contextTx := dbent.TxFromContext(ctx); contextTx != nil {
		return replantCrop(ctx, contextTx.Client(), orderID)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, err
	}
	replanted, err = replantCrop(ctx, tx.Client(), orderID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	return replanted, tx.Commit()
}

func replantCrop(ctx context.Context, client *dbent.Client, orderID int64) (float64, error) {
	rows, err := client.QueryContext(ctx, `
		SELECT id, user_id, amount - stolen_amount FROM garden_crops
		WHERE order_id = $1 AND status = 'removed' FOR UPDATE`, orderID)
	if err != nil {
		return 0, err
	}
	var cropID, userID int64
	var left float64
	found := rows.Next()
	if found {
		err = rows.Scan(&cropID, &userID, &left)
	}
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil || !found {
		return 0, err
	}
	// 余额已经不够盖住这棵菜（用户这段时间花掉了）就不再种回去，不让余额变负。
	taken, err := client.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $2, total_recharged = GREATEST(total_recharged - $2, 0), updated_at = NOW()
		WHERE id = $1 AND balance >= $2`, userID, left)
	if err != nil {
		return 0, err
	}
	if affected, _ := taken.RowsAffected(); affected == 0 {
		return 0, nil
	}
	if _, err = client.ExecContext(ctx, `UPDATE garden_crops SET status = 'growing' WHERE id = $1`, cropID); err != nil {
		return 0, err
	}
	return left, nil
}

func (r *gardenRepository) PlantGift(ctx context.Context, email string, amount float64, ripenAfter time.Duration) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO garden_crops (user_id, source, amount, ripe_at)
		SELECT id, 'gift', $2, NOW() + make_interval(secs => $3)
		FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
		ORDER BY id LIMIT 1
		RETURNING user_id`, email, amount, ripenAfter.Seconds()).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUserNotFound
	}
	return userID, err
}

func (r *gardenRepository) PlantInviteReward(ctx context.Context, inviteeID, orderID int64, amount float64, ripenAfter time.Duration) (int64, error) {
	// 只认被邀请人的第一笔付款；唯一索引保证同一个被邀请人只给邀请人一棵。
	const query = `
		INSERT INTO garden_crops (user_id, source, amount, ripe_at, ref_user_id, ref_order_id)
		SELECT a.inviter_id, 'invite', $3, NOW() + make_interval(secs => $4), $1, $2
		FROM user_affiliates a
		JOIN users inviter ON inviter.id = a.inviter_id AND inviter.deleted_at IS NULL AND inviter.status = 'active'
		WHERE a.user_id = $1 AND a.inviter_id IS NOT NULL
			AND NOT EXISTS (SELECT 1 FROM payment_orders p WHERE p.user_id = $1 AND p.id <> $2 AND p.paid_at IS NOT NULL)
		ON CONFLICT (ref_user_id) WHERE source = 'invite' DO NOTHING
		RETURNING user_id
	`
	var inviterID int64
	err := r.db.QueryRowContext(ctx, query, inviteeID, orderID, amount, ripenAfter.Seconds()).Scan(&inviterID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return inviterID, err
}

func (r *gardenRepository) RemoveInviteReward(ctx context.Context, orderID int64) error {
	// 这棵菜是站里送的，没从谁的余额里扣过，拔掉不用退钱；邻居已经偷走的部分不追。
	_, err := r.db.ExecContext(ctx, `
		UPDATE garden_crops SET status = 'removed'
		WHERE ref_order_id = $1 AND source = 'invite' AND status = 'growing'`, orderID)
	return err
}

func (r *gardenRepository) Harvest(ctx context.Context, userID, cropID int64) (float64, bool, error) {
	const query = `
		WITH harvested AS (
			UPDATE garden_crops
			SET status = 'harvested', harvested_at = NOW(), harvested_amount = amount - stolen_amount
			WHERE id = $1 AND user_id = $2 AND status = 'growing' AND ripe_at <= NOW()
			RETURNING user_id, amount - stolen_amount AS left_amount
		)
		UPDATE users AS u
		SET balance = u.balance + harvested.left_amount,
			total_recharged = u.total_recharged + harvested.left_amount,
			updated_at = NOW()
		FROM harvested
		WHERE u.id = harvested.user_id
		RETURNING harvested.left_amount
	`
	return r.movedAmount(ctx, query, cropID, userID)
}

func (r *gardenRepository) GetCrop(ctx context.Context, userID, cropID int64) (*service.GardenCrop, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+gardenCropColumns+` FROM garden_crops WHERE id = $1 AND user_id = $2`, cropID, userID)
	crop, err := scanGardenCrop(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return crop, err
}

func (r *gardenRepository) ListCrops(ctx context.Context, userID int64, harvestedLimit int) (crops []service.GardenCrop, err error) {
	const query = `
		(SELECT ` + gardenCropColumns + ` FROM garden_crops WHERE user_id = $1 AND status = 'growing')
		UNION ALL
		(SELECT ` + gardenCropColumns + ` FROM garden_crops WHERE user_id = $1 AND status = 'harvested'
			ORDER BY harvested_at DESC LIMIT $2)
		ORDER BY 6, 8, 1
	`
	rows, err := r.db.QueryContext(ctx, query, userID, harvestedLimit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	crops = make([]service.GardenCrop, 0)
	for rows.Next() {
		crop, scanErr := scanGardenCrop(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		crops = append(crops, *crop)
	}
	return crops, rows.Err()
}
