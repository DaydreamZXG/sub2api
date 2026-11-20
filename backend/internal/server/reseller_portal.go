package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/web"
	"github.com/gin-gonic/gin"
)

const resellerPortalContextKey = "reseller_portal"

// A plain dot, so the client domain's tab does not carry the site's own icon when no logo was set.
const resellerPortalBlankIcon = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Ccircle cx='8' cy='8' r='6' fill='%23111'/%3E%3C/svg%3E"

type resellerPortalLookup interface {
	PortalFor(ctx context.Context, host string) (*service.ResellerPortal, error)
}

// resellerPortalFor finds the client domain's settings for a request. A proxy may pass the original
// host on; either header naming the client domain is enough, so that nothing of the site is served
// there by mistake.
func resellerPortalFor(c *gin.Context, lookup resellerPortalLookup) (*service.ResellerPortal, error) {
	if portal, err := lookup.PortalFor(c.Request.Context(), c.Request.Host); portal != nil || err != nil {
		return portal, err
	}
	if forwarded := strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Host"), ",")[0]); forwarded != "" {
		return lookup.PortalFor(c.Request.Context(), forwarded)
	}
	return nil, nil
}

// resellerPortalGuard keeps the resellers' client domain to the gateway and the client pages. The
// site's own API, with its prices, accounts and sign-in, answers 404 there, as do the setup wizard
// and the OAuth discovery documents.
func resellerPortalGuard(lookup resellerPortalLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		portal, err := resellerPortalFor(c, lookup)
		if err != nil {
			// The setting could not be read yet, so this may be the client domain: nothing of the
			// site is served until it is known. Only the health check answers.
			if c.Request.URL.Path == "/health" {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": http.StatusServiceUnavailable, "message": "starting up, try again shortly"})
			return
		}
		if portal == nil {
			c.Next()
			return
		}
		c.Set(resellerPortalContextKey, portal)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || path == "/api" || strings.HasPrefix(path, "/setup") || strings.HasPrefix(path, "/.well-known/") {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "not found"})
			return
		}
		c.Next()
	}
}

// resellerPortalPage gives the frontend server the config for the client pages. sign is how amounts
// are written on the site, so a client reads their budget the way the reseller does.
func resellerPortalPage(sign func(ctx context.Context) string) web.PortalPageFunc {
	return func(c *gin.Context) ([]byte, bool) {
		value, ok := c.Get(resellerPortalContextKey)
		if !ok {
			return nil, false
		}
		portal, ok := value.(*service.ResellerPortal)
		if !ok || portal == nil {
			return nil, false
		}
		icon := portal.Logo
		if icon == "" {
			icon = resellerPortalBlankIcon
		}
		config, err := json.Marshal(map[string]any{
			"portal":       true,
			"site_name":    portal.Name,
			"site_logo":    icon,
			"portal_logo":  portal.Logo,
			"contact_info": portal.Contact,
			"sign":         sign(c.Request.Context()),
		})
		if err != nil {
			return nil, false
		}
		return config, true
	}
}

// resellerPortalSign mirrors the frontend: stored amounts are yuan only while one unit of balance is
// bought for one yuan.
func resellerPortalSign(settings web.PublicSettingsProvider) func(ctx context.Context) string {
	return func(ctx context.Context) string {
		public, err := settings.GetPublicSettingsForInjection(ctx)
		if err != nil {
			return "¥"
		}
		raw, err := json.Marshal(public)
		if err != nil {
			return "¥"
		}
		var cfg struct {
			PaymentEnabled bool    `json:"payment_enabled"`
			Multiplier     float64 `json:"balance_recharge_multiplier"`
		}
		if json.Unmarshal(raw, &cfg) != nil {
			return "¥"
		}
		if cfg.PaymentEnabled && cfg.Multiplier > 0 && cfg.Multiplier != 1 {
			return "$"
		}
		return "¥"
	}
}
