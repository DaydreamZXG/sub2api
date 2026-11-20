package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrowthLevelsLevelAndDiscount(t *testing.T) {
	ladder := DefaultGrowthLevels()
	require.Equal(t, -1, ladder.LevelFor(5000), "the default ladder is off until the owner enables it")
	require.Zero(t, ladder.DiscountFor(5000))

	ladder.Enabled = true
	for _, tc := range []struct {
		total    float64
		level    int
		discount float64
	}{
		{0, 0, 0},
		{99.99, 0, 0},
		{100, 1, 0.02},
		{499, 1, 0.02},
		{2000, 3, 0.08},
		{1e9, 4, 0.12},
	} {
		require.Equal(t, tc.level, ladder.LevelFor(tc.total), "total %v", tc.total)
		require.InDelta(t, tc.discount, ladder.DiscountFor(tc.total), 1e-12, "total %v", tc.total)
	}
}

func TestGrowthLevelsValidate(t *testing.T) {
	valid := DefaultGrowthLevels()
	require.NoError(t, valid.Validate())

	for name, ladder := range map[string]GrowthLevels{
		"no levels":             {},
		"first level above 0":   {Levels: []GrowthLevel{{Name: "a", Threshold: 10}}},
		"thresholds not rising": {Levels: []GrowthLevel{{Name: "a"}, {Name: "b", Threshold: 100}, {Name: "c", Threshold: 100}}},
		"blank name":            {Levels: []GrowthLevel{{Name: "  "}}},
		"negative discount":     {Levels: []GrowthLevel{{Name: "a", Discount: -0.1}}},
		"discount above 90%":    {Levels: []GrowthLevel{{Name: "a", Discount: 0.95}}},
		"too many levels":       {Levels: make([]GrowthLevel, maxGrowthLevels+1)},
	} {
		t.Run(name, func(t *testing.T) {
			require.ErrorIs(t, ladder.Validate(), ErrGrowthLevelsInvalid)
		})
	}
}
