package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"
)

// 桌面助手“一键连接”的站点侧实现，接口形状遵循桌面助手公开的
// station-connect 协议（见 yeschoy-ai-desktop/docs/station-connect-api.openapi.yaml）。

const (
	// DesktopConnectCodeTTL 是一次性连接码的有效期，协议要求不超过 300 秒。
	DesktopConnectCodeTTL = 300 * time.Second
	// AppTokenHeader 是桌面助手转发请求时携带设备凭证的请求头。
	AppTokenHeader = "X-Yeschoy-App-Token"

	appTokenPrefix        = "yca_"
	desktopDeviceIDMaxLen = 128
)

var (
	ErrDesktopOriginInvalid   = errors.New("connect_origin_invalid")
	ErrDesktopCodeExpired     = errors.New("connect_code_expired")
	ErrDesktopCodeAlreadyUsed = errors.New("connect_code_already_used")
	ErrDesktopOriginMismatch  = errors.New("connect_origin_mismatch")
	ErrDesktopDeviceInvalid   = errors.New("device_invalid")
)

// DesktopConnectCode is what a one-time code resolves to.
type DesktopConnectCode struct {
	UserID int64  `json:"user_id"`
	Origin string `json:"origin"`
	// The user's token version when the code was issued: a password change revokes the code
	// together with the session that asked for it.
	TokenVersion int64 `json:"token_version"`
}

// DesktopConnectCodeStore keeps one-time codes; Consume must hand a code to exactly one caller.
type DesktopConnectCodeStore interface {
	Save(ctx context.Context, codeHash string, code DesktopConnectCode, ttl time.Duration) error
	// Consume returns ErrDesktopCodeAlreadyUsed for a code that was consumed before and
	// ErrDesktopCodeExpired for an unknown or expired one.
	Consume(ctx context.Context, codeHash string) (*DesktopConnectCode, error)
}

type AppDevice struct {
	DeviceID      string
	Platform      string
	ClientVersion string
}

type AppDeviceTokenRepository interface {
	// Replace revokes the device's active tokens and stores the new hash, in one transaction.
	Replace(ctx context.Context, userID int64, device AppDevice, tokenHash string) error
	// ActiveUserID returns the owner of a token that is not revoked; ok is false when there is none.
	ActiveUserID(ctx context.Context, tokenHash string) (userID int64, ok bool, err error)
}

type DesktopIssuedCode struct {
	Code      string
	Origin    string
	ExpiresAt time.Time
}

type DesktopAuthBundle struct {
	AccessToken      string
	RefreshToken     string
	SessionID        string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type DesktopConnectService struct {
	codes       DesktopConnectCodeStore
	appTokens   AppDeviceTokenRepository
	userRepo    UserRepository
	authService *AuthService
}

func NewDesktopConnectService(codes DesktopConnectCodeStore, appTokens AppDeviceTokenRepository, userRepo UserRepository, authService *AuthService) *DesktopConnectService {
	return &DesktopConnectService{codes: codes, appTokens: appTokens, userRepo: userRepo, authService: authService}
}

// CanonicalDesktopOrigin validates an origin against the protocol: https only, no userinfo,
// no explicit port, no path/query/fragment, and a DNS host name rather than an IP literal.
func CanonicalDesktopOrigin(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" ||
		parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", ErrDesktopOriginInvalid
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" || net.ParseIP(host) != nil {
		return "", ErrDesktopOriginInvalid
	}
	return "https://" + host, nil
}

func randomURLSafeToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashDesktopSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// IssueCode creates a one-time code for the signed-in user, bound to the station origin.
func (s *DesktopConnectService) IssueCode(ctx context.Context, userID int64, origin string) (*DesktopIssuedCode, error) {
	canonical, err := CanonicalDesktopOrigin(origin)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	code, err := randomURLSafeToken()
	if err != nil {
		return nil, err
	}
	issued := DesktopConnectCode{UserID: userID, Origin: canonical, TokenVersion: resolvedTokenVersion(user)}
	if err := s.codes.Save(ctx, hashDesktopSecret(code), issued, DesktopConnectCodeTTL); err != nil {
		return nil, err
	}
	return &DesktopIssuedCode{Code: code, Origin: canonical, ExpiresAt: time.Now().Add(DesktopConnectCodeTTL)}, nil
}

// Exchange turns a one-time code into a normal sign-in session for the desktop app.
// requestHost is the Host the app actually called; it must be the host the code was issued for.
func (s *DesktopConnectService) Exchange(ctx context.Context, code, requestHost string) (*DesktopAuthBundle, error) {
	issued, err := s.codes.Consume(ctx, hashDesktopSecret(code))
	if err != nil {
		return nil, err
	}
	host := strings.ToLower(requestHost)
	if h, _, splitErr := net.SplitHostPort(host); splitErr == nil {
		host = h
	}
	if "https://"+host != issued.Origin {
		return nil, ErrDesktopOriginMismatch
	}
	user, err := s.userRepo.GetByID(ctx, issued.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}
	if issued.TokenVersion != resolvedTokenVersion(user) {
		return nil, ErrDesktopCodeExpired
	}
	sessionBytes := make([]byte, 16)
	if _, err := rand.Read(sessionBytes); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(sessionBytes)
	pair, err := s.authService.GenerateTokenPair(ctx, user, sessionID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &DesktopAuthBundle{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		SessionID:        sessionID,
		AccessExpiresAt:  now.Add(time.Duration(pair.ExpiresIn) * time.Second),
		RefreshExpiresAt: now.Add(s.authService.RefreshTokenTTL()),
	}, nil
}

// IssueAppToken gives the signed-in app on one device its token, replacing the device's previous one.
func (s *DesktopConnectService) IssueAppToken(ctx context.Context, userID int64, device AppDevice) (string, error) {
	device.DeviceID = strings.TrimSpace(device.DeviceID)
	if device.DeviceID == "" || len(device.DeviceID) > desktopDeviceIDMaxLen {
		return "", ErrDesktopDeviceInvalid
	}
	switch device.Platform {
	case "macos", "windows", "linux":
	default:
		device.Platform = "unknown"
	}
	if len(device.ClientVersion) > 32 {
		device.ClientVersion = device.ClientVersion[:32]
	}
	secret, err := randomURLSafeToken()
	if err != nil {
		return "", err
	}
	token := appTokenPrefix + secret
	if err := s.appTokens.Replace(ctx, userID, device, hashDesktopSecret(token)); err != nil {
		return "", err
	}
	return token, nil
}

// IsAppRequest reports whether token is a live app token that belongs to userID.
func (s *DesktopConnectService) IsAppRequest(ctx context.Context, token string, userID int64) bool {
	return NewAppRequestVerifier(s.appTokens).IsAppRequest(ctx, token, userID)
}

// RefreshTokenTTL is how long a refresh token issued now stays valid.
func (s *AuthService) RefreshTokenTTL() time.Duration {
	return time.Duration(s.cfg.JWT.RefreshTokenExpireDays) * 24 * time.Hour
}
