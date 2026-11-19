package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeGardenRepo struct {
	GardenRepository
	planted    map[int64]float64
	ripenAfter time.Duration
	crop       *GardenCrop
}

func (r *fakeGardenRepo) Plant(_ context.Context, _, orderID int64, amount float64, ripenAfter time.Duration) (bool, error) {
	if _, exists := r.planted[orderID]; exists {
		return false, nil
	}
	r.planted[orderID] = amount
	r.ripenAfter = ripenAfter
	return true, nil
}

func (r *fakeGardenRepo) Harvest(_ context.Context, _, _ int64) (float64, bool, error) {
	if r.crop != nil && r.crop.Status == GardenCropGrowing && !r.crop.RipeAt.After(time.Now()) {
		return r.crop.Amount - r.crop.StolenAmount, true, nil
	}
	return 0, false, nil
}

func (r *fakeGardenRepo) GetCrop(context.Context, int64, int64) (*GardenCrop, error) {
	return r.crop, nil
}

func newTestGarden(settings string) (*GardenService, *fakeGardenRepo) {
	repo := &fakeGardenRepo{planted: map[int64]float64{}}
	return NewGardenService(repo, &growthLevelsSettingRepo{value: settings}, nil), repo
}

// The bonus is planted only while the garden is on, once per order, with the configured ripening time.
func TestGardenPlantForOrder(t *testing.T) {
	ctx := context.Background()

	off, offRepo := newTestGarden("")
	planted, err := off.PlantForOrder(ctx, 1, 10, 5)
	require.NoError(t, err)
	require.False(t, planted)
	require.Empty(t, offRepo.planted)

	on, onRepo := newTestGarden(`{"enabled":true,"ripen_hours":12}`)
	planted, err = on.PlantForOrder(ctx, 1, 10, 5)
	require.NoError(t, err)
	require.True(t, planted)
	require.Equal(t, 12*time.Hour, onRepo.ripenAfter)

	planted, err = on.PlantForOrder(ctx, 1, 10, 5)
	require.NoError(t, err)
	require.False(t, planted, "an order is planted once")

	planted, err = on.PlantForOrder(ctx, 1, 11, 0)
	require.NoError(t, err)
	require.False(t, planted, "an order without a bonus plants nothing")
	require.Equal(t, map[int64]float64{10: 5}, onRepo.planted)
}

func TestGardenHarvest(t *testing.T) {
	ctx := context.Background()
	garden, repo := newTestGarden(`{"enabled":true,"ripen_hours":24}`)

	_, err := garden.Harvest(ctx, 1, 1)
	require.ErrorIs(t, err, ErrGardenCropNotFound)

	repo.crop = &GardenCrop{ID: 1, Amount: 8, StolenAmount: 2, Status: GardenCropGrowing, RipeAt: time.Now().Add(time.Hour)}
	_, err = garden.Harvest(ctx, 1, 1)
	require.ErrorIs(t, err, ErrGardenCropNotRipe)

	repo.crop.RipeAt = time.Now().Add(-time.Minute)
	amount, err := garden.Harvest(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, 6.0, amount, "the owner gets what was not stolen")

	repo.crop.Status = GardenCropHarvested
	_, err = garden.Harvest(ctx, 1, 1)
	require.ErrorIs(t, err, ErrGardenCropNotFound, "a crop is harvested once")
}

func TestGardenSettings(t *testing.T) {
	ctx := context.Background()
	garden, _ := newTestGarden("")
	require.Equal(t, DefaultGardenSettings(), garden.Settings(ctx))

	for _, hours := range []float64{0, -1, 721} {
		_, err := garden.SetSettings(ctx, GardenSettings{Enabled: true, RipenHours: hours})
		require.ErrorIs(t, err, ErrGardenSettingsInvalid)
	}
	saved, err := garden.SetSettings(ctx, GardenSettings{Enabled: true, RipenHours: 6})
	require.NoError(t, err)
	require.Equal(t, saved, garden.Settings(ctx))

	unreadable, _ := newTestGarden(`{"enabled":true,"ripen_hours":0}`)
	require.False(t, unreadable.Settings(ctx).Enabled, "unusable saved rules mean the garden is off")
}
