package service

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeGardenRepo struct {
	GardenRepository
	planted    map[int64]float64
	ripenAfter time.Duration
	crop       *GardenCrop
	growing    float64
	// gifts are {user, amount} of crops the site gave; inviters maps an invitee to their inviter.
	gifts      [][2]float64
	inviters   map[int64]int64
	rewarded   map[int64]bool
	growingErr error
}

func (r *fakeGardenRepo) PlantGift(_ context.Context, _ string, amount float64, _ time.Duration) (int64, error) {
	r.gifts = append(r.gifts, [2]float64{9, amount})
	return 9, nil
}

func (r *fakeGardenRepo) PlantInviteReward(_ context.Context, inviteeID, _ int64, amount float64, _ time.Duration) (int64, error) {
	if r.inviters[inviteeID] == 0 || r.rewarded[inviteeID] {
		return 0, nil
	}
	r.rewarded[inviteeID] = true
	r.gifts = append(r.gifts, [2]float64{float64(r.inviters[inviteeID]), amount})
	return r.inviters[inviteeID], nil
}

func (r *fakeGardenRepo) RemoveInviteReward(context.Context, int64) error { return nil }

func (r *fakeGardenRepo) GrowingForOrder(context.Context, int64) (float64, error) {
	return r.growing, r.growingErr
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
	return NewGardenService(repo, nil, &growthLevelsSettingRepo{value: settings}, nil, nil), repo
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

// A crop the site gives away needs someone to give it to and an amount within the limit.
func TestGardenGiftCrop(t *testing.T) {
	ctx := context.Background()
	repo := &fakeGardenRepo{planted: map[int64]float64{}}
	garden := NewGardenService(repo, nil, &growthLevelsSettingRepo{}, nil, nil)

	for _, bad := range []struct {
		email  string
		amount float64
	}{{"", 5}, {"a@example.com", 0}, {"a@example.com", -1}, {"a@example.com", GardenMaxGiftAmount + 1}, {"a@example.com", math.NaN()}} {
		_, err := garden.GiftCrop(ctx, bad.email, bad.amount)
		require.ErrorIs(t, err, ErrGardenGiftInvalid)
	}
	require.Empty(t, repo.gifts)

	userID, err := garden.GiftCrop(ctx, " a@example.com ", 5)
	require.NoError(t, err, "a gift does not need the garden to be open")
	require.Equal(t, int64(9), userID)
	require.Equal(t, [][2]float64{{9, 5}}, repo.gifts)
}

// The inviter gets one crop per invited user, and only while the garden is open and the reward is set.
func TestGardenInviteReward(t *testing.T) {
	ctx := context.Background()
	newGarden := func(settings string) (*GardenService, *fakeGardenRepo) {
		repo := &fakeGardenRepo{planted: map[int64]float64{}, inviters: map[int64]int64{2: 1}, rewarded: map[int64]bool{}}
		return NewGardenService(repo, nil, &growthLevelsSettingRepo{value: settings}, nil, nil), repo
	}

	for _, off := range []string{
		`{"enabled":false,"ripen_hours":24,"invite_crop_amount":5}`,
		`{"enabled":true,"ripen_hours":24}`,
	} {
		garden, repo := newGarden(off)
		planted, err := garden.PlantInviteReward(ctx, 2, 100)
		require.NoError(t, err)
		require.False(t, planted)
		require.Empty(t, repo.gifts)
	}

	garden, repo := newGarden(`{"enabled":true,"ripen_hours":24,"invite_crop_amount":5}`)
	planted, err := garden.PlantInviteReward(ctx, 2, 100)
	require.NoError(t, err)
	require.True(t, planted)
	planted, err = garden.PlantInviteReward(ctx, 2, 101)
	require.NoError(t, err)
	require.False(t, planted, "a second order of the same invitee earns nothing")
	planted, err = garden.PlantInviteReward(ctx, 3, 102)
	require.NoError(t, err)
	require.False(t, planted, "nobody invited user 3")
	require.Equal(t, [][2]float64{{1, 5}}, repo.gifts)
}
