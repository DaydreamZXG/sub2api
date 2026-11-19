package repository

import (
	"context"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/appdevicetoken"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type appDeviceTokenRepository struct {
	client *dbent.Client
}

func NewAppDeviceTokenRepository(client *dbent.Client) service.AppDeviceTokenRepository {
	return &appDeviceTokenRepository{client: client}
}

func (r *appDeviceTokenRepository) Replace(ctx context.Context, userID int64, device service.AppDevice, tokenHash string) error {
	replace := func(client *dbent.Client) error {
		err := client.AppDeviceToken.Update().
			Where(appdevicetoken.UserIDEQ(userID), appdevicetoken.DeviceIDEQ(device.DeviceID), appdevicetoken.RevokedAtIsNil()).
			SetRevokedAt(time.Now()).
			Exec(ctx)
		if err != nil {
			return err
		}
		return client.AppDeviceToken.Create().
			SetUserID(userID).
			SetDeviceID(device.DeviceID).
			SetPlatform(device.Platform).
			SetClientVersion(device.ClientVersion).
			SetTokenHash(tokenHash).
			Exec(ctx)
	}
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return replace(tx.Client())
	}
	tx, err := r.client.Tx(ctx)
	if errors.Is(err, dbent.ErrTxStarted) {
		return replace(r.client)
	}
	if err != nil {
		return err
	}
	if err := replace(tx.Client()); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *appDeviceTokenRepository) ActiveUserID(ctx context.Context, tokenHash string) (int64, bool, error) {
	token, err := clientFromContext(ctx, r.client).AppDeviceToken.Query().
		Where(appdevicetoken.TokenHashEQ(tokenHash), appdevicetoken.RevokedAtIsNil()).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return token.UserID, true, nil
}
