package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type gardenRepository struct {
	db *sql.DB
}

// NewGardenRepository creates the garden store.
func NewGardenRepository(db *sql.DB) service.GardenRepository {
	return &gardenRepository{db: db}
}

const gardenCropColumns = `id, user_id, COALESCE(order_id, 0), amount, stolen_amount, status, planted_at, ripe_at, harvested_at, harvested_amount`

func scanGardenCrop(row interface{ Scan(dest ...any) error }) (*service.GardenCrop, error) {
	var (
		crop        service.GardenCrop
		harvestedAt sql.NullTime
	)
	if err := row.Scan(&crop.ID, &crop.UserID, &crop.OrderID, &crop.Amount, &crop.StolenAmount, &crop.Status,
		&crop.PlantedAt, &crop.RipeAt, &harvestedAt, &crop.HarvestedAmount); err != nil {
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

func (r *gardenRepository) Replant(ctx context.Context, orderID int64) (replanted float64, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var cropID, userID int64
	var left float64
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, amount - stolen_amount FROM garden_crops
		WHERE order_id = $1 AND status = 'removed' FOR UPDATE`, orderID).Scan(&cropID, &userID, &left)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, tx.Rollback()
	}
	if err != nil {
		return 0, err
	}
	// 余额已经不够盖住这棵菜（用户这段时间花掉了）就不再种回去，不让余额变负。
	taken, err := tx.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $2, total_recharged = GREATEST(total_recharged - $2, 0), updated_at = NOW()
		WHERE id = $1 AND balance >= $2`, userID, left)
	if err != nil {
		return 0, err
	}
	if rows, _ := taken.RowsAffected(); rows == 0 {
		return 0, tx.Rollback()
	}
	if _, err = tx.ExecContext(ctx, `UPDATE garden_crops SET status = 'growing' WHERE id = $1`, cropID); err != nil {
		return 0, err
	}
	return left, tx.Commit()
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
