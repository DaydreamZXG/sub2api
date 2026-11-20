package handler

import (
	"context"
	"log/slog"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type siteStatsSource interface {
	GetDashboardStats(ctx context.Context) (*usagestats.DashboardStats, error)
}

// SiteStatsHandler serves the site-wide usage totals shown on the public home page.
type SiteStatsHandler struct {
	source siteStatsSource
}

// NewSiteStatsHandler reads the totals from the cached admin dashboard statistics.
func NewSiteStatsHandler(dashboardService *service.DashboardService) *SiteStatsHandler {
	return &SiteStatsHandler{source: dashboardService}
}

// siteStatsResponse holds counts only: no users, accounts or money.
type siteStatsResponse struct {
	TotalTokens   int64 `json:"total_tokens"`
	TodayTokens   int64 `json:"today_tokens"`
	TotalRequests int64 `json:"total_requests"`
}

// Get returns the site-wide token and request totals. No sign-in required.
func (h *SiteStatsHandler) Get(c *gin.Context) {
	stats, err := h.source.GetDashboardStats(c.Request.Context())
	if err != nil {
		slog.Error("site stats: load dashboard stats failed", "error", err)
		response.InternalError(c, "Failed to load site statistics")
		return
	}
	response.Success(c, siteStatsResponse{
		TotalTokens:   stats.TotalTokens,
		TodayTokens:   stats.TodayTokens,
		TotalRequests: stats.TotalRequests,
	})
}
