package service

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

type appLineGroupRepo struct {
	GroupRepository
	groups map[int64]*Group
}

func (r appLineGroupRepo) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	if group, ok := r.groups[id]; ok {
		return group, nil
	}
	return nil, ErrGroupNotFound
}

func TestResolveAppLine(t *testing.T) {
	ctx := context.Background()
	const ownerID, strangerID = int64(7), int64(8)
	normalID, missingID := int64(2), int64(99)

	normal := &Group{ID: normalID, Name: "normal", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, Platform: PlatformAnthropic, RateMultiplier: 1}
	disabled := &Group{ID: 3, Status: StatusDisabled, SubscriptionType: SubscriptionTypeStandard}
	subscription := &Group{ID: 4, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}
	otherAppLine := &Group{ID: 5, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, AppOnly: true}
	otherPlatform := &Group{ID: 6, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, Platform: PlatformOpenAI}

	tokens := newFakeAppDeviceTokenRepo()
	connect := NewDesktopConnectService(newFakeDesktopCodeStore(), tokens, nil, nil)
	ownerToken, err := connect.IssueAppToken(ctx, ownerID, AppDevice{DeviceID: "owner-device"})
	require.NoError(t, err)
	strangerToken, err := connect.IssueAppToken(ctx, strangerID, AppDevice{DeviceID: "stranger-device"})
	require.NoError(t, err)

	svc := &APIKeyService{groupRepo: appLineGroupRepo{groups: map[int64]*Group{
		normal.ID: normal, disabled.ID: disabled, subscription.ID: subscription, otherAppLine.ID: otherAppLine, otherPlatform.ID: otherPlatform,
	}}}
	svc.SetAppRequestVerifier(NewAppRequestVerifier(tokens))

	keyOn := func(fallback *int64) *APIKey {
		appLine := &Group{ID: 1, Name: "app", Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, Platform: PlatformAnthropic,
			RateMultiplier: 0.8, AppOnly: true, AppFallbackGroupID: fallback}
		return &APIKey{ID: 11, User: &User{ID: ownerID}, GroupID: &appLine.ID, Group: appLine}
	}

	t.Run("the owner's app token keeps the app line", func(t *testing.T) {
		key := keyOn(&normalID)
		resolved, err := svc.ResolveAppLine(ctx, key, ownerToken)
		require.NoError(t, err)
		require.Same(t, key, resolved)
	})

	for name, token := range map[string]string{"no token": "", "made-up token": "yca_made-up", "another user's token": strangerToken} {
		t.Run(name+" is served and billed by the normal-price line", func(t *testing.T) {
			key := keyOn(&normalID)
			resolved, err := svc.ResolveAppLine(ctx, key, token)
			require.NoError(t, err)
			require.Equal(t, normalID, *resolved.GroupID)
			require.Same(t, normal, resolved.Group)
			require.Equal(t, int64(1), key.Group.ID, "the cached key must not be modified")
			require.Equal(t, int64(1), *key.GroupID)
		})
	}

	t.Run("without a normal-price line the request is refused", func(t *testing.T) {
		_, err := svc.ResolveAppLine(ctx, keyOn(nil), "")
		require.ErrorIs(t, err, ErrAppOnlyLine)
	})

	for name, id := range map[string]int64{"deleted": missingID, "disabled": disabled.ID, "subscription": subscription.ID, "app-only": otherAppLine.ID, "other-platform": otherPlatform.ID} {
		t.Run("a "+name+" fallback is refused rather than used", func(t *testing.T) {
			fallbackID := id
			_, err := svc.ResolveAppLine(ctx, keyOn(&fallbackID), "")
			require.ErrorIs(t, err, ErrAppOnlyLine)
		})
	}

	t.Run("a line without the switch is left alone", func(t *testing.T) {
		key := &APIKey{ID: 12, User: &User{ID: ownerID}, GroupID: &normal.ID, Group: normal}
		resolved, err := svc.ResolveAppLine(ctx, key, "")
		require.NoError(t, err)
		require.Same(t, key, resolved)
	})

	t.Run("without a verifier nothing counts as the app", func(t *testing.T) {
		bare := &APIKeyService{groupRepo: svc.groupRepo}
		_, err := bare.ResolveAppLine(ctx, keyOn(nil), ownerToken)
		require.ErrorIs(t, err, ErrAppOnlyLine)
	})
}

type growthLevelsSettingRepo struct {
	SettingRepository
	value string
}

func (r *growthLevelsSettingRepo) GetValue(context.Context, string) (string, error) {
	if r.value == "" {
		return "", ErrSettingNotFound
	}
	return r.value, nil
}

func (r *growthLevelsSettingRepo) Set(_ context.Context, _ string, value string) error {
	r.value = value
	return nil
}

// The growth level takes its discount off the app line only, and only for requests made through the app.
func TestResolveAppLine_GrowthLevelDiscount(t *testing.T) {
	ctx := context.Background()
	const ownerID = int64(7)
	normalID := int64(2)
	normal := &Group{ID: normalID, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, Platform: PlatformAnthropic, RateMultiplier: 1}

	tokens := newFakeAppDeviceTokenRepo()
	connect := NewDesktopConnectService(newFakeDesktopCodeStore(), tokens, nil, nil)
	ownerToken, err := connect.IssueAppToken(ctx, ownerID, AppDevice{DeviceID: "owner-device"})
	require.NoError(t, err)

	levels := NewGrowthLevelsService(&growthLevelsSettingRepo{})
	svc := &APIKeyService{groupRepo: appLineGroupRepo{groups: map[int64]*Group{normal.ID: normal}}}
	svc.SetAppRequestVerifier(NewAppRequestVerifier(tokens))
	svc.SetGrowthLevels(levels)

	keyFor := func(totalRecharged float64) *APIKey {
		appLine := &Group{ID: 1, Status: StatusActive, SubscriptionType: SubscriptionTypeStandard, Platform: PlatformAnthropic,
			RateMultiplier: 0.8, AppOnly: true, AppFallbackGroupID: &normalID}
		return &APIKey{ID: 11, User: &User{ID: ownerID, TotalRecharged: totalRecharged}, GroupID: &appLine.ID, Group: appLine}
	}

	t.Run("no discount while the ladder is off", func(t *testing.T) {
		key := keyFor(5000)
		resolved, err := svc.ResolveAppLine(ctx, key, ownerToken)
		require.NoError(t, err)
		require.Same(t, key, resolved)
	})

	ladder := DefaultGrowthLevels()
	ladder.Enabled = true
	_, err = levels.Set(ctx, ladder)
	require.NoError(t, err)

	t.Run("the app price drops by the level's discount", func(t *testing.T) {
		key := keyFor(500) // 长叶: 5% off
		resolved, err := svc.ResolveAppLine(ctx, key, ownerToken)
		require.NoError(t, err)
		require.InDelta(t, 0.76, resolved.Group.RateMultiplier, 1e-9)
		require.Equal(t, int64(1), *resolved.GroupID)
		require.InDelta(t, 0.8, key.Group.RateMultiplier, 1e-12, "the cached key and its group are never modified")
	})

	t.Run("the first level changes nothing", func(t *testing.T) {
		key := keyFor(50)
		resolved, err := svc.ResolveAppLine(ctx, key, ownerToken)
		require.NoError(t, err)
		require.Same(t, key, resolved)
	})

	t.Run("requests outside the app pay the normal line's price, whatever the level", func(t *testing.T) {
		resolved, err := svc.ResolveAppLine(ctx, keyFor(10000), "")
		require.NoError(t, err)
		require.Same(t, normal, resolved.Group)
		require.InDelta(t, 1.0, resolved.Group.RateMultiplier, 1e-12)
	})

	t.Run("an invalid ladder is refused and the saved one stays", func(t *testing.T) {
		_, err := levels.Set(ctx, GrowthLevels{Enabled: true})
		require.ErrorIs(t, err, ErrGrowthLevelsInvalid)
		require.InDelta(t, 0.05, levels.Get(ctx).DiscountFor(500), 1e-12)
	})
}

func TestDiscountedAppRate(t *testing.T) {
	cases := []struct {
		rate, discount, want float64
	}{
		{rate: 0.8, discount: 0.1, want: 0.72},
		{rate: 1, discount: 0.33333, want: 0.6667},
		{rate: 0, discount: 0.5, want: 0},
		{rate: 1, discount: 1, want: 0},
		// 0.0001 with 90% off is 0.00001: it must stay a price, not round down to free.
		{rate: 0.0001, discount: 0.9, want: 0.00001},
	}
	for _, tc := range cases {
		got := discountedAppRate(tc.rate, tc.discount)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("discountedAppRate(%v, %v) = %v, want %v", tc.rate, tc.discount, got, tc.want)
		}
	}
}
