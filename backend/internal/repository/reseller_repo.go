package repository

import (
	"context"
	"database/sql"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type resellerRepository struct {
	db *sql.DB
}

// NewResellerRepository creates the store for who is a reseller and what their client keys used.
func NewResellerRepository(db *sql.DB) service.ResellerRepository {
	return &resellerRepository{db: db}
}

func (r *resellerRepository) IsReseller(ctx context.Context, userID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM resellers WHERE user_id = $1)`, userID).Scan(&ok)
	return ok, err
}

func (r *resellerRepository) SetReseller(ctx context.Context, userID int64, on bool) (bool, error) {
	if !on {
		// A reseller whose account was deleted since can still be taken off the list.
		res, err := r.db.ExecContext(ctx, `DELETE FROM resellers WHERE user_id = $1`, userID)
		if err != nil {
			return false, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			return true, nil
		}
		var exists bool
		err = r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`, userID).Scan(&exists)
		return exists, err
	}
	var found bool
	err := r.db.QueryRowContext(ctx, `
		WITH u AS (SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL),
		added AS (INSERT INTO resellers (user_id) SELECT id FROM u ON CONFLICT (user_id) DO NOTHING)
		SELECT EXISTS (SELECT 1 FROM u)`, userID).Scan(&found)
	return found, err
}

// A client key's cost to the reseller is what it used at the client's price divided by its markup:
// every call adds exactly cost times markup to quota_used.
func (r *resellerRepository) ListResellers(ctx context.Context) ([]service.ResellerEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rs.user_id, COALESCE(u.email, ''), rs.created_at, COALESCE(u.balance, 0),
			COUNT(k.id), COALESCE(SUM(k.quota_used), 0), COALESCE(SUM(k.quota_used / k.client_markup), 0)
		FROM resellers rs
		JOIN users u ON u.id = rs.user_id AND u.deleted_at IS NULL
		LEFT JOIN api_keys k ON k.user_id = rs.user_id AND k.client_markup > 0 AND k.deleted_at IS NULL
		GROUP BY rs.user_id, u.email, rs.created_at, u.balance
		ORDER BY rs.created_at DESC, rs.user_id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []service.ResellerEntry
	for rows.Next() {
		var e service.ResellerEntry
		if err := rows.Scan(&e.UserID, &e.Email, &e.Since, &e.Balance, &e.Keys, &e.ClientUsed, &e.Cost); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *resellerRepository) ClientKeys(ctx context.Context, userID int64) ([]service.ResellerClientKey, float64, error) {
	var balance float64
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(balance, 0) FROM users WHERE id = $1`, userID).Scan(&balance); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT k.id, k.name, k.key, k.group_id, COALESCE(g.name, ''), k.status, k.client_markup,
			k.quota, k.quota_used, k.expires_at, k.last_used_at, k.created_at
		FROM api_keys k
		LEFT JOIN groups g ON g.id = k.group_id
		WHERE k.user_id = $1 AND k.client_markup > 0 AND k.deleted_at IS NULL
		ORDER BY k.created_at DESC, k.id DESC`, userID)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var out []service.ResellerClientKey
	for rows.Next() {
		var (
			k        service.ResellerClientKey
			groupID  sql.NullInt64
			expires  sql.NullTime
			lastUsed sql.NullTime
		)
		if err := rows.Scan(&k.ID, &k.Name, &k.Key, &groupID, &k.GroupName, &k.Status, &k.Markup,
			&k.Budget, &k.ClientUsed, &expires, &lastUsed, &k.CreatedAt); err != nil {
			return nil, 0, err
		}
		if groupID.Valid {
			k.GroupID = &groupID.Int64
		}
		if expires.Valid {
			k.ExpiresAt = &expires.Time
		}
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		out = append(out, k)
	}
	return out, balance, rows.Err()
}
