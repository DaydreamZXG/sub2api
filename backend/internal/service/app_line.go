package service

import (
	"context"
	"errors"
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
		return apiKey, nil
	}
	if apiKey.Group.AppFallbackGroupID == nil {
		return nil, ErrAppOnlyLine
	}
	fallback, err := s.groupRepo.GetByIDLite(ctx, *apiKey.Group.AppFallbackGroupID)
	if errors.Is(err, ErrGroupNotFound) || (err == nil && !AppFallbackGroupUsable(fallback)) {
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
