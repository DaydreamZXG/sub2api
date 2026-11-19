package service

import (
	"context"
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
