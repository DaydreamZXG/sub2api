package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterSiteStatsRoutes registers the public, read-only site usage totals.
func RegisterSiteStatsRoutes(v1 *gin.RouterGroup, h *handler.Handlers, panelRateLimiter *middleware.PanelRateLimiter) {
	v1.GET("/site-stats", panelRateLimiter.PublicIP(), h.SiteStats.Get)
}
