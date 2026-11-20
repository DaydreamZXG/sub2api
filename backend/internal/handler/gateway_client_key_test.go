package handler

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestKeyBillingInfoIsNotForAClientKey(t *testing.T) {
	groupID := int64(7)
	apiKey := &service.APIKey{
		UserID:       11,
		GroupID:      &groupID,
		ClientMarkup: 1.5,
		Group:        &service.Group{ID: groupID, Name: "private-group-name", RateMultiplier: 0.75},
	}
	c, w := newKeyBillingContext(apiKey)

	newKeyBillingHandler(nil).KeyBillingInfo(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.NotContains(t, w.Body.String(), "0.75")
	require.NotContains(t, w.Body.String(), "multiplier")
}

func TestClientPriceUsagePutsEveryAmountInTheClientsPrice(t *testing.T) {
	summary := gin.H{
		"today": gin.H{"requests": int64(2), "cost": 10.0, "actual_cost": 4.0},
		"total": gin.H{"requests": int64(9), "cost": 50.0, "actual_cost": 20.0},
		"rpm":   int64(3),
	}
	clientPriceUsageSummary(summary, 1.5)
	require.Equal(t, gin.H{"requests": int64(2), "cost": 6.0, "actual_cost": 6.0}, summary["today"])
	require.Equal(t, gin.H{"requests": int64(9), "cost": 30.0, "actual_cost": 30.0}, summary["total"])
	require.Equal(t, int64(3), summary["rpm"])

	stored := []usagestats.APIKeyDailyUsagePoint{{Date: "2026-11-20", Requests: 2, TotalTokens: 100, Cost: 10, ActualCost: 4}}
	daily := clientPriceDailyUsage(stored, 1.5)
	require.Equal(t, []usagestats.APIKeyDailyUsagePoint{{Date: "2026-11-20", Requests: 2, TotalTokens: 100, Cost: 6, ActualCost: 6}}, daily)
	// The figures the caller got are left alone.
	require.Equal(t, 4.0, stored[0].ActualCost)

	models := clientPriceModelStats([]usagestats.ModelStat{{Model: "m", Requests: 2, Cost: 10, ActualCost: 4, AccountCost: 3}}, 1.5)
	require.Equal(t, []usagestats.ModelStat{{Model: "m", Requests: 2, Cost: 6, ActualCost: 6, AccountCost: 0}}, models)
}
