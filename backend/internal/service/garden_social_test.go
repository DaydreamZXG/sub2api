package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeGardenSocialRepo struct {
	GardenSocialRepository
	neighbours []GardenNeighbour
	crops      []GardenNeighbourCropState
	ownerID    int64
	ownerTotal float64
	stealRules *GardenStealRules
	feed       []GardenFeedItem
	seenUntil  []time.Time
}

func (r *fakeGardenSocialRepo) Feed(context.Context, int64, int) ([]GardenFeedItem, error) {
	return r.feed, nil
}

func (r *fakeGardenSocialRepo) MarkFeedSeen(_ context.Context, _ int64, until time.Time) error {
	r.seenUntil = append(r.seenUntil, until)
	return nil
}

func (r *fakeGardenSocialRepo) Neighbours(context.Context, int64, int) ([]GardenNeighbour, error) {
	return append([]GardenNeighbour(nil), r.neighbours...), nil
}

func (r *fakeGardenSocialRepo) NeighbourCrops(context.Context, int64, int64) ([]GardenNeighbourCropState, error) {
	return r.crops, nil
}

func (r *fakeGardenSocialRepo) CropOwner(context.Context, int64) (int64, float64, bool, error) {
	return r.ownerID, r.ownerTotal, r.ownerID != 0, nil
}

func (r *fakeGardenSocialRepo) Steal(_ context.Context, _, _ int64, rules GardenStealRules) (GardenStealResult, error) {
	r.stealRules = &rules
	return GardenStealResult{Caught: rules.DogCatches}, nil
}

func newTestSocialGarden(settings string) (*GardenService, *fakeGardenSocialRepo) {
	social := &fakeGardenSocialRepo{}
	garden := NewGardenService(&fakeGardenRepo{planted: map[int64]float64{}}, social, &growthLevelsSettingRepo{value: settings}, nil, nil)
	return garden, social
}

func TestGardenName(t *testing.T) {
	require.Equal(t, "老张", gardenName(" 老张 ", "zhang@example.com"))
	require.Equal(t, "zh***@example.com", gardenName("", "zhang@example.com"))
	require.Equal(t, "a***@example.com", gardenName("", "a@example.com"))
	require.Equal(t, "***", gardenName("", "not-an-email"))
}

// A stranger sees a crop's size and what they may do to it, never its amount.
func TestGardenNeighbourGarden(t *testing.T) {
	ctx := context.Background()
	garden, social := newTestSocialGarden(`{"enabled":true,"ripen_hours":24}`)
	social.neighbours = []GardenNeighbour{{ID: 2, Email: "bob@example.com", Crops: 4}}
	now := time.Now()
	social.crops = []GardenNeighbourCropState{
		{ID: 1, Amount: 5, RipeAt: now.Add(3 * time.Hour)},                       // growing: can be watered
		{ID: 2, Amount: 50, RipeAt: now.Add(-time.Hour)},                         // ripe, still protected
		{ID: 3, Amount: 500, RipeAt: now.Add(-3 * time.Hour)},                    // open
		{ID: 4, Amount: 500, StolenAmount: 250, RipeAt: now.Add(-3 * time.Hour)}, // open, picked clean
		{ID: 5, Amount: 500, RipeAt: now.Add(-3 * time.Hour), Tried: "caught"},   // already tried
		{ID: 6, Amount: 5, RipeAt: now.Add(3 * time.Hour), Waters: 5},            // watered enough
		{ID: 7, Amount: 5, RipeAt: now.Add(3 * time.Hour), WateredToday: true},   // watered by the viewer today
	}

	neighbour, crops, err := garden.NeighbourGarden(ctx, 1, 2)
	require.NoError(t, err)
	require.Equal(t, "bo***@example.com", neighbour.Name)
	type seen struct {
		size            string
		water, stealing bool
	}
	got := make([]seen, 0, len(crops))
	for _, crop := range crops {
		got = append(got, seen{crop.Size, crop.CanWater, crop.CanSteal})
	}
	require.Equal(t, []seen{
		{"small", true, false},
		{"medium", false, false},
		{"large", false, true},
		{"large", false, false},
		{"large", false, false},
		{"small", false, false},
		{"small", false, false},
	}, got)

	_, _, err = garden.NeighbourGarden(ctx, 1, 3)
	require.ErrorIs(t, err, ErrGardenNotNeighbour)
}

func TestGardenStealChecksBeforeTheTry(t *testing.T) {
	ctx := context.Background()

	closed, _ := newTestSocialGarden("")
	_, err := closed.Steal(ctx, 1, 10)
	require.ErrorIs(t, err, ErrGardenClosed)

	garden, social := newTestSocialGarden(`{"enabled":true,"ripen_hours":24}`)
	_, err = garden.Steal(ctx, 1, 10)
	require.ErrorIs(t, err, ErrGardenCropNotFound)

	social.ownerID = 1
	_, err = garden.Steal(ctx, 1, 10)
	require.ErrorIs(t, err, ErrGardenOwnCrop)

	social.ownerID = 2
	_, err = garden.Steal(ctx, 1, 10)
	require.ErrorIs(t, err, ErrGardenNotNeighbour)
	require.Nil(t, social.stealRules, "nothing is tried in a garden that is not a neighbour's today")

	social.neighbours = []GardenNeighbour{{ID: 2}}
	garden.rollDice = func() float64 { return 0.05 } // under the 10% of a dog whose owner has no growth level
	result, err := garden.Steal(ctx, 1, 10)
	require.NoError(t, err)
	require.True(t, result.Caught)
	require.Equal(t, GardenStealRules{
		ProtectFor: 2 * time.Hour, StealShare: 0.03, MaxStolenShare: 0.5, DailyLimit: 10, DogCatches: true,
	}, *social.stealRules)

	garden.rollDice = func() float64 { return 0.5 }
	result, err = garden.Steal(ctx, 1, 10)
	require.NoError(t, err)
	require.False(t, result.Caught)
}

// Reading the feed marks as read only what neighbours did that the list actually shows.
func TestGardenFeedMarksOnlyWhatItShows(t *testing.T) {
	ctx := context.Background()
	garden, social := newTestSocialGarden(`{"enabled":true,"ripen_hours":24}`)
	base := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)

	social.feed = []GardenFeedItem{{Kind: GardenActionSteal, Mine: true, CreatedAt: base.Add(3 * time.Hour)}}
	_, err := garden.Feed(ctx, 1)
	require.NoError(t, err)
	require.Empty(t, social.seenUntil, "the user's own actions are not something to read")

	social.feed = []GardenFeedItem{
		{Kind: GardenActionSteal, Mine: true, CreatedAt: base.Add(3 * time.Hour)},
		{Kind: GardenActionWater, Email: "bob@example.com", CreatedAt: base.Add(2 * time.Hour)},
		{Kind: GardenActionCaught, Email: "bob@example.com", CreatedAt: base.Add(time.Hour)},
	}
	items, err := garden.Feed(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "bo***@example.com", items[1].Name)
	require.Equal(t, []time.Time{base.Add(2 * time.Hour)}, social.seenUntil)
}

// The dog of an owner on a higher growth level catches thieves more often.
func TestGardenDogChanceFollowsGrowthLevel(t *testing.T) {
	ctx := context.Background()
	garden, _ := newTestSocialGarden("")
	settings := DefaultGardenSettings()
	require.InDelta(t, 0.10, garden.dogChance(ctx, settings, 99999), 1e-9, "without growth levels every dog is the weakest")

	garden.growthLevels = NewGrowthLevelsService(&growthLevelsSettingRepo{
		value: `{"enabled":true,"levels":[{"name":"a","threshold":0,"discount":0},{"name":"b","threshold":100,"discount":0.1},{"name":"c","threshold":1000,"discount":0.2}]}`,
	})
	require.InDelta(t, 0.10, garden.dogChance(ctx, settings, 0), 1e-9)
	require.InDelta(t, 0.30, garden.dogChance(ctx, settings, 100), 1e-9)
	require.InDelta(t, 0.50, garden.dogChance(ctx, settings, 5000), 1e-9)
}
