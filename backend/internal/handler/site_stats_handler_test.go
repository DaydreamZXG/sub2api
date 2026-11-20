package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type siteStatsSourceStub struct {
	stats *usagestats.DashboardStats
	err   error
}

func (s siteStatsSourceStub) GetDashboardStats(context.Context) (*usagestats.DashboardStats, error) {
	return s.stats, s.err
}

func serveSiteStats(t *testing.T, source siteStatsSource) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/site-stats", (&SiteStatsHandler{source: source}).Get)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/site-stats", nil))
	return recorder
}

func TestSiteStatsGet_ReturnsOnlyUsageCounts(t *testing.T) {
	recorder := serveSiteStats(t, siteStatsSourceStub{stats: &usagestats.DashboardStats{
		TotalUsers:      42,
		TotalRequests:   1200,
		TotalTokens:     9000000,
		TodayTokens:     350000,
		TotalCost:       88.5,
		TotalActualCost: 61.2,
	}})

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t,
		`{"code":0,"message":"success","data":{"total_tokens":9000000,"today_tokens":350000,"total_requests":1200}}`,
		recorder.Body.String())
}

func TestSiteStatsGet_SourceFailureIsServerError(t *testing.T) {
	recorder := serveSiteStats(t, siteStatsSourceStub{err: errors.New("cache down")})

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "cache down")
}
