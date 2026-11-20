package service

import (
	"context"
	"errors"
	"math"
	"strings"
)

// 种菜线路：仅限野菜 app 的分组。请求带着本人设备的 app 凭证时按该分组计费；
// 否则改用分组指定的普通价分组，没有指定则拒绝。

// ErrAppOnlyLine means the key's line is reserved for the app and has no usable fallback.
var ErrAppOnlyLine = errors.New("this line only works through the YESCHOY app")

// AppRequestVerifier tells whether a request was forwarded by the signed-in app of this user.
type AppRequestVerifier interface {
	IsAppRequest(ctx context.Context, token string, userID int64) bool
}

type appTokenVerifier struct {
	appTokens AppDeviceTokenRepository
}

// NewAppRequestVerifier recognises app requests by the per-device token issued through the connect flow.
func NewAppRequestVerifier(appTokens AppDeviceTokenRepository) AppRequestVerifier {
	return appTokenVerifier{appTokens: appTokens}
}

func (v appTokenVerifier) IsAppRequest(ctx context.Context, token string, userID int64) bool {
	if !strings.HasPrefix(token, appTokenPrefix) {
		return false
	}
	owner, ok, err := v.appTokens.ActiveUserID(ctx, hashDesktopSecret(token))
	return err == nil && ok && owner == userID
}

// SetAppRequestVerifier enables app-only lines; without it every request counts as not from the app.
func (s *APIKeyService) SetAppRequestVerifier(verifier AppRequestVerifier) {
	s.appRequestVerifier = verifier
}

// SetGrowthLevels enables the growth-level discount on app-only lines.
func (s *APIKeyService) SetGrowthLevels(levels *GrowthLevelsService) {
	s.growthLevels = levels
}

// AppLineGroupForUser returns the app-only group as this user is billed for it through the app:
// a copy with the growth-level discount taken off the price, or group itself when no discount applies.
// A price an admin set for this user on the group still wins, because billing looks that up afterwards.
func (s *APIKeyService) AppLineGroupForUser(ctx context.Context, group *Group, user *User) *Group {
	if group == nil || !group.AppOnly || user == nil || s.growthLevels == nil {
		return group
	}
	discount := s.growthLevels.Get(ctx).DiscountFor(user.TotalRecharged)
	if discount <= 0 {
		return group
	}
	discounted := *group
	discounted.RateMultiplier = discountedAppRate(group.RateMultiplier, discount)
	return &discounted
}

// discountedAppRate takes the discount off a line's price, rounded to four decimals.
// A price too small for four decimals is left as it is: rounding it would make the line free.
func discountedAppRate(rate, discount float64) float64 {
	exact := rate * (1 - discount)
	rounded := math.Round(exact*1e4) / 1e4
	if rounded == 0 && exact > 0 {
		return exact
	}
	return rounded
}

// HasManualAppLinePrice reports whether an admin set this user's own price on every app-only line they can use,
// so the growth-level discount applies to none of them.
func (s *APIKeyService) HasManualAppLinePrice(ctx context.Context, userID int64) bool {
	rates, err := s.GetUserGroupRates(ctx, userID)
	if err != nil || len(rates) == 0 {
		return false
	}
	groups, err := s.GetAvailableGroups(ctx, userID)
	if err != nil {
		return false
	}
	appLines := 0
	for i := range groups {
		if !groups[i].AppOnly {
			continue
		}
		if _, manual := rates[groups[i].ID]; !manual {
			return false
		}
		appLines++
	}
	return appLines > 0
}

// AppLineRatesForUser returns the growth-level price of each app-only group in groups for this user,
// keyed by group ID. Groups without a discount are left out. It is for display: billing uses AppLineGroupForUser.
func (s *APIKeyService) AppLineRatesForUser(ctx context.Context, userID int64, groups []*Group) map[int64]float64 {
	if s.growthLevels == nil || !s.growthLevels.Get(ctx).Enabled {
		return nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil
	}
	rates := map[int64]float64{}
	for _, group := range groups {
		if discounted := s.AppLineGroupForUser(ctx, group, user); discounted != group {
			rates[group.ID] = discounted.RateMultiplier
		}
	}
	return rates
}

// AppFallbackGroupUsable reports whether group can serve requests that an app-only line turns away.
func AppFallbackGroupUsable(group *Group) bool {
	return group != nil && group.IsActive() && !group.IsSubscriptionType() && !group.AppOnly
}

// ResolveAppLine returns the key to serve this request with. Everything downstream (scheduling,
// billing, usage logs) reads the key's group, so swapping it here is the only place the price changes.
// The returned key is a copy when the group was swapped; the cached key is never modified.
func (s *APIKeyService) ResolveAppLine(ctx context.Context, apiKey *APIKey, appToken string) (*APIKey, error) {
	if apiKey == nil || apiKey.Group == nil || !apiKey.Group.AppOnly {
		return apiKey, nil
	}
	if apiKey.User != nil && s.appRequestVerifier != nil && s.appRequestVerifier.IsAppRequest(ctx, appToken, apiKey.User.ID) {
		group := s.AppLineGroupForUser(ctx, apiKey.Group, apiKey.User)
		if group == apiKey.Group {
			return apiKey, nil
		}
		resolved := *apiKey
		resolved.Group = group
		return &resolved, nil
	}
	if apiKey.Group.AppFallbackGroupID == nil {
		return nil, ErrAppOnlyLine
	}
	fallback, err := s.groupRepo.GetByIDLite(ctx, *apiKey.Group.AppFallbackGroupID)
	if errors.Is(err, ErrGroupNotFound) || (err == nil && (!AppFallbackGroupUsable(fallback) || fallback.Platform != apiKey.Group.Platform)) {
		return nil, ErrAppOnlyLine
	}
	if err != nil {
		return nil, err
	}
	resolved := *apiKey
	resolved.GroupID = &fallback.ID
	resolved.Group = fallback
	return &resolved, nil
}
