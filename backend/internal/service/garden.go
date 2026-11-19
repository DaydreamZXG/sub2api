package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync/atomic"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// SettingKeyGarden stores the garden rules as JSON in the settings table.
const SettingKeyGarden = "garden_settings"

const (
	gardenSettingsCacheTTL = 60 * time.Second
	gardenMaxRipenHours    = 720
	gardenHarvestedShown   = 20

	GardenCropGrowing   = "growing"
	GardenCropHarvested = "harvested"
	GardenCropRemoved   = "removed"
)

var (
	ErrGardenSettingsInvalid = infraerrors.BadRequest("GARDEN_SETTINGS_INVALID", "invalid garden settings")
	ErrGardenCropNotFound    = infraerrors.NotFound("GARDEN_CROP_NOT_FOUND", "crop not found")
	ErrGardenCropNotRipe     = infraerrors.Conflict("GARDEN_CROP_NOT_RIPE", "crop is not ripe yet")
)

// GardenSettings are the garden rules an admin can change.
// While Enabled, the bonus part of a balance top-up is planted as a crop instead of being credited at once.
type GardenSettings struct {
	Enabled    bool    `json:"enabled"`
	RipenHours float64 `json:"ripen_hours"`
	// ProtectHours: after ripening, only the owner can touch the crop for this long.
	ProtectHours float64 `json:"protect_hours"`
	// StealPercent of a crop goes to each successful thief; MaxStolenPercent is the most a crop can lose in total.
	StealPercent     float64 `json:"steal_percent"`
	MaxStolenPercent float64 `json:"max_stolen_percent"`
	DailyStealLimit  int     `json:"daily_steal_limit"`
	// Each watering by a neighbour ripens the crop WaterMinutes sooner, at most MaxWatersPerCrop times.
	WaterMinutes     float64 `json:"water_minutes"`
	MaxWatersPerCrop int     `json:"max_waters_per_crop"`
	NeighboursPerDay int     `json:"neighbours_per_day"`
	// A fed dog catches a thief with a chance between DogMinPercent (lowest growth level) and DogMaxPercent (highest).
	DogMinPercent float64 `json:"dog_min_percent"`
	DogMaxPercent float64 `json:"dog_max_percent"`
}

// DefaultGardenSettings is the garden switched off with a one-day ripening time.
func DefaultGardenSettings() GardenSettings {
	return GardenSettings{
		Enabled: false, RipenHours: 24, ProtectHours: 2,
		StealPercent: 3, MaxStolenPercent: 50, DailyStealLimit: 10,
		WaterMinutes: 60, MaxWatersPerCrop: 5, NeighboursPerDay: 10,
		DogMinPercent: 10, DogMaxPercent: 50,
	}
}

// Validate reports whether the rules can be saved.
func (g GardenSettings) Validate() error {
	if math.IsNaN(g.RipenHours) || g.RipenHours <= 0 || g.RipenHours > gardenMaxRipenHours {
		return ErrGardenSettingsInvalid
	}
	within := func(value, low, high float64) bool { return !math.IsNaN(value) && value >= low && value <= high }
	switch {
	case !within(g.ProtectHours, 0, gardenMaxRipenHours),
		!within(g.StealPercent, 0, 100), !within(g.MaxStolenPercent, 0, 100), g.StealPercent > g.MaxStolenPercent,
		!within(g.WaterMinutes, 0, 24*60),
		!within(g.DogMinPercent, 0, 100), !within(g.DogMaxPercent, 0, 100), g.DogMinPercent > g.DogMaxPercent,
		g.DailyStealLimit < 0 || g.DailyStealLimit > 1000,
		g.MaxWatersPerCrop < 0 || g.MaxWatersPerCrop > 100,
		g.NeighboursPerDay < 0 || g.NeighboursPerDay > 50:
		return ErrGardenSettingsInvalid
	}
	return nil
}

// GardenCrop is one planted top-up bonus.
type GardenCrop struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"-"`
	OrderID         int64      `json:"-"`
	Amount          float64    `json:"amount"`
	StolenAmount    float64    `json:"stolen_amount"`
	Status          string     `json:"status"`
	PlantedAt       time.Time  `json:"planted_at"`
	RipeAt          time.Time  `json:"ripe_at"`
	HarvestedAt     *time.Time `json:"harvested_at,omitempty"`
	HarvestedAmount float64    `json:"harvested_amount"`
}

// GardenRepository moves a crop's value between the crop and the user's balance, each in one statement.
type GardenRepository interface {
	// Plant takes amount out of the user's balance into a new crop for the order; false when the order already has one.
	Plant(ctx context.Context, userID, orderID int64, amount float64, ripenAfter time.Duration) (bool, error)
	// Uproot returns what is left of the order's growing crop to the balance and reports that amount.
	Uproot(ctx context.Context, orderID int64) (float64, error)
	// Harvest moves a ripe crop into its owner's balance; ok is false when no ripe growing crop matched.
	Harvest(ctx context.Context, userID, cropID int64) (amount float64, ok bool, err error)
	GetCrop(ctx context.Context, userID, cropID int64) (*GardenCrop, error)
	ListCrops(ctx context.Context, userID int64, harvestedLimit int) ([]GardenCrop, error)
}

type cachedGardenSettings struct {
	settings  GardenSettings
	expiresAt time.Time
}

// GardenService plants top-up bonuses as crops and lets their owners harvest them.
type GardenService struct {
	repo          GardenRepository
	settingRepo   SettingRepository
	redeemService *RedeemService
	cache         atomic.Pointer[cachedGardenSettings]
}

// NewGardenService creates the garden.
func NewGardenService(repo GardenRepository, settingRepo SettingRepository, redeemService *RedeemService) *GardenService {
	return &GardenService{repo: repo, settingRepo: settingRepo, redeemService: redeemService}
}

// Settings returns the saved rules, or the garden switched off when none are saved or they cannot be read.
func (s *GardenService) Settings(ctx context.Context) GardenSettings {
	if s == nil || s.settingRepo == nil {
		return DefaultGardenSettings()
	}
	if cached := s.cache.Load(); cached != nil && time.Now().Before(cached.expiresAt) {
		return cached.settings
	}
	settings := DefaultGardenSettings()
	raw, err := s.settingRepo.GetValue(context.WithoutCancel(ctx), SettingKeyGarden)
	if err == nil && strings.TrimSpace(raw) != "" {
		// 旧版本存的设置没有后加的字段，从默认值上叠加。
		saved := DefaultGardenSettings()
		if json.Unmarshal([]byte(raw), &saved) == nil && saved.Validate() == nil {
			settings = saved
		}
	}
	s.cache.Store(&cachedGardenSettings{settings: settings, expiresAt: time.Now().Add(gardenSettingsCacheTTL)})
	return settings
}

// SetSettings validates and saves the rules. Crops already planted keep their ripening time.
func (s *GardenService) SetSettings(ctx context.Context, settings GardenSettings) (GardenSettings, error) {
	if err := settings.Validate(); err != nil {
		return GardenSettings{}, err
	}
	blob, err := json.Marshal(settings)
	if err != nil {
		return GardenSettings{}, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyGarden, string(blob)); err != nil {
		return GardenSettings{}, err
	}
	s.cache.Store(&cachedGardenSettings{settings: settings, expiresAt: time.Now().Add(gardenSettingsCacheTTL)})
	return settings, nil
}

// PlantForOrder moves a fulfilled top-up's bonus out of the balance into a crop.
// It does nothing while the garden is off or the order has no bonus, and at most once per order.
func (s *GardenService) PlantForOrder(ctx context.Context, userID, orderID int64, bonus float64) (bool, error) {
	if s == nil || s.repo == nil || bonus <= 0 || math.IsNaN(bonus) || math.IsInf(bonus, 0) {
		return false, nil
	}
	settings := s.Settings(ctx)
	if !settings.Enabled {
		return false, nil
	}
	planted, err := s.repo.Plant(ctx, userID, orderID, bonus, time.Duration(settings.RipenHours*float64(time.Hour)))
	if err != nil {
		return false, err
	}
	if planted {
		s.invalidateBalance(ctx, userID)
	}
	return planted, nil
}

// UprootForOrder puts what is left of a refunded order's growing crop back into the balance,
// so the refund then deducts from the same balance it would have without the garden.
func (s *GardenService) UprootForOrder(ctx context.Context, userID, orderID int64) (float64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	returned, err := s.repo.Uproot(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if returned > 0 {
		s.invalidateBalance(ctx, userID)
	}
	return returned, nil
}

// Crops lists the user's growing crops and their most recent harvests.
func (s *GardenService) Crops(ctx context.Context, userID int64) ([]GardenCrop, error) {
	return s.repo.ListCrops(ctx, userID, gardenHarvestedShown)
}

// Harvest moves a ripe crop into its owner's balance. It works even after the garden is switched off.
func (s *GardenService) Harvest(ctx context.Context, userID, cropID int64) (float64, error) {
	amount, ok, err := s.repo.Harvest(ctx, userID, cropID)
	if err != nil {
		return 0, err
	}
	if ok {
		s.invalidateBalance(ctx, userID)
		return amount, nil
	}
	crop, err := s.repo.GetCrop(ctx, userID, cropID)
	if err != nil {
		return 0, err
	}
	if crop == nil || crop.Status != GardenCropGrowing {
		return 0, ErrGardenCropNotFound
	}
	return 0, ErrGardenCropNotRipe
}

func (s *GardenService) invalidateBalance(ctx context.Context, userID int64) {
	if s.redeemService == nil {
		return
	}
	s.redeemService.invalidateRedeemCaches(ctx, userID, &RedeemCode{Type: RedeemTypeBalance})
}

// IsGardenRuleError reports whether err is a rule violation the caller caused, not a failure.
func IsGardenRuleError(err error) bool {
	return errors.Is(err, ErrGardenCropNotFound) || errors.Is(err, ErrGardenCropNotRipe) || errors.Is(err, ErrGardenSettingsInvalid)
}
