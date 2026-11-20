package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

// 种菜计划的成长等级：累计充值越多，仅限 app 的线路价格越低。等级只看累计充值，
// 累计充值只增不减，所以等级只升不降。

// SettingKeyGrowthLevels stores the GrowthLevels JSON.
const SettingKeyGrowthLevels = "growth_levels"

const maxGrowthLevels = 10

// GrowthLevel is one step: reached once the user's cumulative top-ups hit Threshold.
type GrowthLevel struct {
	Name      string  `json:"name"`
	Threshold float64 `json:"threshold"` // cumulative top-ups, in balance units
	Discount  float64 `json:"discount"`  // share taken off the app line's price, 0-0.9
}

// GrowthLevels is the whole ladder; disabled means nobody gets a level discount.
type GrowthLevels struct {
	Enabled bool          `json:"enabled"`
	Levels  []GrowthLevel `json:"levels"`
}

// DefaultGrowthLevels is an example ladder for the owner to edit; it is off until enabled.
func DefaultGrowthLevels() GrowthLevels {
	return GrowthLevels{Levels: []GrowthLevel{
		{Name: "种子", Threshold: 0, Discount: 0},
		{Name: "发芽", Threshold: 100, Discount: 0.02},
		{Name: "长叶", Threshold: 500, Discount: 0.05},
		{Name: "开花", Threshold: 2000, Discount: 0.08},
		{Name: "丰收", Threshold: 10000, Discount: 0.12},
	}}
}

// ErrGrowthLevelsInvalid is returned for a ladder the admin API must refuse.
var ErrGrowthLevelsInvalid = errors.New("growth levels: need 1-10 named levels, thresholds rising from 0, discounts between 0 and 0.9")

// Validate checks the ladder and trims level names.
func (g *GrowthLevels) Validate() error {
	if len(g.Levels) == 0 || len(g.Levels) > maxGrowthLevels || g.Levels[0].Threshold != 0 {
		return ErrGrowthLevelsInvalid
	}
	for i := range g.Levels {
		level := &g.Levels[i]
		level.Name = strings.TrimSpace(level.Name)
		if level.Name == "" || math.IsNaN(level.Threshold) || math.IsInf(level.Threshold, 0) ||
			math.IsNaN(level.Discount) || level.Discount < 0 || level.Discount > 0.9 {
			return ErrGrowthLevelsInvalid
		}
		if i > 0 && level.Threshold <= g.Levels[i-1].Threshold {
			return ErrGrowthLevelsInvalid
		}
	}
	return nil
}

// LevelFor returns the index of the highest level reached with totalRecharged, or -1 when the ladder is off.
func (g GrowthLevels) LevelFor(totalRecharged float64) int {
	if !g.Enabled || len(g.Levels) == 0 {
		return -1
	}
	reached := 0
	for i, level := range g.Levels {
		if totalRecharged >= level.Threshold {
			reached = i
		}
	}
	return reached
}

// DiscountFor returns the share taken off the app line's price for this user, 0 when the ladder is off.
func (g GrowthLevels) DiscountFor(totalRecharged float64) float64 {
	index := g.LevelFor(totalRecharged)
	if index < 0 {
		return 0
	}
	return g.Levels[index].Discount
}

const growthLevelsCacheTTL = 60 * time.Second

type cachedGrowthLevels struct {
	levels    GrowthLevels
	expiresAt time.Time
}

// GrowthLevelsService stores the ladder in the settings table and serves it to the request path from a short cache.
type GrowthLevelsService struct {
	settingRepo SettingRepository
	cache       atomic.Pointer[cachedGrowthLevels]
}

// NewGrowthLevelsService creates the ladder store.
func NewGrowthLevelsService(settingRepo SettingRepository) *GrowthLevelsService {
	return &GrowthLevelsService{settingRepo: settingRepo}
}

// Get returns the saved ladder, or the disabled example ladder when none is saved or it cannot be read.
// Other instances pick up a change within growthLevelsCacheTTL.
func (s *GrowthLevelsService) Get(ctx context.Context) GrowthLevels {
	if s == nil || s.settingRepo == nil {
		return DefaultGrowthLevels()
	}
	if cached := s.cache.Load(); cached != nil && time.Now().Before(cached.expiresAt) {
		return cached.levels.clone()
	}
	levels := DefaultGrowthLevels()
	raw, err := s.settingRepo.GetValue(context.WithoutCancel(ctx), SettingKeyGrowthLevels)
	if err == nil && strings.TrimSpace(raw) != "" {
		var saved GrowthLevels
		if json.Unmarshal([]byte(raw), &saved) == nil && saved.Validate() == nil {
			levels = saved
		}
	}
	// 读库失败时按“未开启”处理：宁可少给折扣，也不让请求失败。
	s.cache.Store(&cachedGrowthLevels{levels: levels.clone(), expiresAt: time.Now().Add(growthLevelsCacheTTL)})
	return levels
}

// Set validates and saves the ladder.
func (s *GrowthLevelsService) Set(ctx context.Context, levels GrowthLevels) (GrowthLevels, error) {
	if err := levels.Validate(); err != nil {
		return GrowthLevels{}, err
	}
	blob, err := json.Marshal(levels)
	if err != nil {
		return GrowthLevels{}, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyGrowthLevels, string(blob)); err != nil {
		return GrowthLevels{}, err
	}
	s.cache.Store(&cachedGrowthLevels{levels: levels.clone(), expiresAt: time.Now().Add(growthLevelsCacheTTL)})
	return levels, nil
}

func (g GrowthLevels) clone() GrowthLevels {
	g.Levels = append([]GrowthLevel(nil), g.Levels...)
	return g
}
