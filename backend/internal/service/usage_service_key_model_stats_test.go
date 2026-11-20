package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

// Embeds the interface: only the one call this test needs is answered.
type keyModelStatsRepoStub struct {
	UsageLogRepository
	stats    []usagestats.ModelStat
	apiKeyID int64
}

func (r *keyModelStatsRepoStub) GetModelStatsWithFilters(_ context.Context, _, _ time.Time, _, apiKeyID, _, _ int64, _ *int16, _ *bool, _ *int8) ([]usagestats.ModelStat, error) {
	r.apiKeyID = apiKeyID
	return r.stats, nil
}

func TestGetAPIKeyModelStatsLeavesOutWhatTheCallsCostUpstream(t *testing.T) {
	repo := &keyModelStatsRepoStub{stats: []usagestats.ModelStat{
		{Model: "a", Requests: 2, Cost: 10, ActualCost: 4, AccountCost: 3},
		{Model: "b", Requests: 1, Cost: 5, ActualCost: 2, AccountCost: 1.5},
	}}
	svc := &UsageService{usageRepo: repo}

	stats, err := svc.GetAPIKeyModelStats(context.Background(), 8, time.Now().Add(-time.Hour), time.Now())

	require.NoError(t, err)
	require.Equal(t, int64(8), repo.apiKeyID)
	require.Equal(t, []usagestats.ModelStat{
		{Model: "a", Requests: 2, Cost: 10, ActualCost: 4},
		{Model: "b", Requests: 1, Cost: 5, ActualCost: 2},
	}, stats)
}
