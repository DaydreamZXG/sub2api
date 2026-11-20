package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/gin-gonic/gin"
)

// A reseller's client holds the key, so whatever the key can ask about itself must be in the client's
// price: what the calls cost the reseller, the list price and the upstream cost are the reseller's
// and the site's business. These helpers rewrite the money in GET /v1/usage for such a key.

func clientPriceUsageSummary(data gin.H, markup float64) {
	for _, period := range []string{"today", "total"} {
		figures, ok := data[period].(gin.H)
		if !ok {
			continue
		}
		actual, _ := figures["actual_cost"].(float64)
		figures["cost"] = actual * markup
		figures["actual_cost"] = actual * markup
	}
}

func clientPriceDailyUsage(points []usagestats.APIKeyDailyUsagePoint, markup float64) []usagestats.APIKeyDailyUsagePoint {
	out := make([]usagestats.APIKeyDailyUsagePoint, len(points))
	for i, p := range points {
		p.Cost = p.ActualCost * markup
		p.ActualCost = p.Cost
		out[i] = p
	}
	return out
}

func clientPriceModelStats(stats []usagestats.ModelStat, markup float64) []usagestats.ModelStat {
	out := make([]usagestats.ModelStat, len(stats))
	for i, s := range stats {
		s.Cost = s.ActualCost * markup
		s.ActualCost = s.Cost
		s.AccountCost = 0
		out[i] = s
	}
	return out
}
