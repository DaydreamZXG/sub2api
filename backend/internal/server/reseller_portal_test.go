package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type portalLookupStub struct {
	portal service.ResellerPortal
	err    error
}

func (s portalLookupStub) PortalFor(_ context.Context, host string) (*service.ResellerPortal, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.portal.Domain == "" || service.NormalizeResellerHost(host) != s.portal.Domain {
		return nil, nil
	}
	portal := s.portal
	return &portal, nil
}

type portalSettingsStub struct {
	value any
	err   error
}

func (s portalSettingsStub) GetPublicSettingsForInjection(context.Context) (any, error) {
	return s.value, s.err
}

func newPortalTestRouter(portal service.ResellerPortal) *gin.Engine {
	return newPortalTestRouterWith(portalLookupStub{portal: portal})
}

func newPortalTestRouterWith(lookup portalLookupStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(resellerPortalGuard(lookup))
	ok := func(c *gin.Context) {
		config, onPortal := resellerPortalPage(func(context.Context) string { return "¥" })(c)
		c.Header("X-On-Portal", map[bool]string{true: "yes", false: "no"}[onPortal])
		c.Data(http.StatusOK, "application/json", append([]byte(nil), config...))
	}
	for _, path := range []string{"/api/v1/settings/public", "/api/v1/auth/login", "/setup/status", "/.well-known/oauth-authorization-server", "/v1/usage", "/v1/messages", "/", "/assets/app.js", "/health", "/api/v3/contents/generations/tasks"} {
		r.Any(path, ok)
	}
	return r
}

func portalTestGet(r *gin.Engine, host, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = host
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestResellerPortalGuardLeavesTheSiteItselfAlone(t *testing.T) {
	r := newPortalTestRouter(service.ResellerPortal{Domain: "keys.example.com", Name: "Keys"})
	for _, path := range []string{"/api/v1/settings/public", "/api/v1/auth/login", "/setup/status", "/v1/usage", "/"} {
		w := portalTestGet(r, "main.example.com", path, nil)
		require.Equal(t, http.StatusOK, w.Code, path)
		require.Equal(t, "no", w.Header().Get("X-On-Portal"), path)
	}

	// With no client domain saved, nothing is one.
	off := newPortalTestRouter(service.ResellerPortal{})
	require.Equal(t, http.StatusOK, portalTestGet(off, "main.example.com", "/api/v1/settings/public", nil).Code)
}

func TestResellerPortalGuardServesOnlyTheGatewayAndThePagesOnTheClientDomain(t *testing.T) {
	r := newPortalTestRouter(service.ResellerPortal{Domain: "keys.example.com", Name: "Keys", Contact: "wx: keys"})
	for _, path := range []string{"/api/v1/settings/public", "/api/v1/auth/login", "/api/v3", "/api/v30/x", "/setup/status", "/.well-known/oauth-authorization-server"} {
		w := portalTestGet(r, "keys.example.com", path, nil)
		require.Equal(t, http.StatusNotFound, w.Code, path)
		require.NotContains(t, w.Body.String(), "portal", path)
	}
	for _, path := range []string{"/v1/usage", "/v1/messages", "/", "/assets/app.js", "/api/v3/contents/generations/tasks"} {
		w := portalTestGet(r, "KEYS.example.com:443", path, nil)
		require.Equal(t, http.StatusOK, w.Code, path)
		require.Equal(t, "yes", w.Header().Get("X-On-Portal"), path)
	}

	// A proxy that keeps its own Host but passes the visitor's on is still the client domain.
	w := portalTestGet(r, "127.0.0.1:8080", "/api/v1/settings/public", map[string]string{"X-Forwarded-Host": "keys.example.com, proxy.internal"})
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestResellerPortalPageCarriesOnlyTheClientDomainsOwnSettings(t *testing.T) {
	r := newPortalTestRouter(service.ResellerPortal{Domain: "keys.example.com", Name: "Keys", Contact: "wx: keys"})
	var config map[string]any
	require.NoError(t, json.Unmarshal(portalTestGet(r, "keys.example.com", "/", nil).Body.Bytes(), &config))
	require.Equal(t, map[string]any{
		"portal": true, "site_name": "Keys", "site_logo": resellerPortalBlankIcon, "portal_logo": "", "contact_info": "wx: keys", "sign": "¥",
	}, config)

	withLogo := newPortalTestRouter(service.ResellerPortal{Domain: "keys.example.com", Name: "Keys", Logo: "https://keys.example.com/logo.png"})
	require.NoError(t, json.Unmarshal(portalTestGet(withLogo, "keys.example.com", "/", nil).Body.Bytes(), &config))
	require.Equal(t, "https://keys.example.com/logo.png", config["site_logo"])
	require.Equal(t, "https://keys.example.com/logo.png", config["portal_logo"])
}

func TestResellerPortalSignFollowsHowTheSiteWritesAmounts(t *testing.T) {
	sign := func(value any, err error) string {
		return resellerPortalSign(portalSettingsStub{value: value, err: err})(context.Background())
	}
	require.Equal(t, "¥", sign(map[string]any{"payment_enabled": false, "balance_recharge_multiplier": 7}, nil))
	require.Equal(t, "¥", sign(map[string]any{"payment_enabled": true, "balance_recharge_multiplier": 1}, nil))
	require.Equal(t, "$", sign(map[string]any{"payment_enabled": true, "balance_recharge_multiplier": 0.14}, nil))
	require.Equal(t, "¥", sign(nil, errors.New("down")))
}

func TestResellerPortalGuardServesNothingOfTheSiteWhileTheDomainIsNotKnown(t *testing.T) {
	r := newPortalTestRouterWith(portalLookupStub{err: errors.New("database is starting")})
	for _, path := range []string{"/api/v1/settings/public", "/", "/v1/usage"} {
		w := portalTestGet(r, "keys.example.com", path, nil)
		require.Equal(t, http.StatusServiceUnavailable, w.Code, path)
	}
	require.Equal(t, http.StatusOK, portalTestGet(r, "keys.example.com", "/health", nil).Code)
}
