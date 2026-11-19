package service

import (
	"context"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	gardenFeedShown = 30
	// Strangers see a crop's size, never its amount.
	gardenMediumCrop = 10
	gardenLargeCrop  = 100

	GardenActionWater  = "water"
	GardenActionSteal  = "steal"
	GardenActionCaught = "caught"
)

var (
	ErrGardenClosed         = infraerrors.Forbidden("GARDEN_CLOSED", "the garden is closed")
	ErrGardenNotNeighbour   = infraerrors.Forbidden("GARDEN_NOT_NEIGHBOUR", "not one of today's neighbours")
	ErrGardenOwnCrop        = infraerrors.BadRequest("GARDEN_OWN_CROP", "this is your own crop")
	ErrGardenCropProtected  = infraerrors.Conflict("GARDEN_CROP_PROTECTED", "only the owner can touch this crop for now")
	ErrGardenNeedCrop       = infraerrors.Conflict("GARDEN_NEED_CROP", "plant a crop of your own first")
	ErrGardenStealLimit     = infraerrors.Conflict("GARDEN_STEAL_LIMIT", "no tries left today")
	ErrGardenAlreadyTried   = infraerrors.Conflict("GARDEN_ALREADY_TRIED", "already tried this crop")
	ErrGardenPickedClean    = infraerrors.Conflict("GARDEN_PICKED_CLEAN", "nothing left to take from this crop")
	ErrGardenCropRipe       = infraerrors.Conflict("GARDEN_CROP_RIPE", "this crop is already ripe")
	ErrGardenWaterFull      = infraerrors.Conflict("GARDEN_WATER_FULL", "this crop has been watered enough")
	ErrGardenAlreadyWatered = infraerrors.Conflict("GARDEN_ALREADY_WATERED", "already watered this crop today")
)

// GardenNeighbour is another user's garden as the viewer may see it.
type GardenNeighbour struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Pinned neighbours are the viewer's inviter and invitees; they stay every day.
	Pinned bool `json:"pinned"`
	// Raided neighbours took, or tried to take, from the viewer in the last 24 hours; they are listed so the viewer can go back.
	Raided bool `json:"took_from_you"`
	Crops  int  `json:"crops"`
	// Username and Email are only used to build Name.
	Username string `json:"-"`
	Email    string `json:"-"`
}

// GardenNeighbourCropState is a neighbour's growing crop with what the viewer already did to it.
type GardenNeighbourCropState struct {
	ID           int64
	Amount       float64
	StolenAmount float64
	PlantedAt    time.Time
	RipeAt       time.Time
	Waters       int
	WateredToday bool
	// Tried is "steal" or "caught" once the viewer tried to take from the crop, else empty.
	Tried string
}

// GardenNeighbourCrop is what the viewer is shown of a neighbour's crop.
type GardenNeighbourCrop struct {
	ID        int64     `json:"id"`
	Size      string    `json:"size"`
	PlantedAt time.Time `json:"planted_at"`
	RipeAt    time.Time `json:"ripe_at"`
	// OpenAt is when neighbours may start taking from it.
	OpenAt   time.Time `json:"open_at"`
	CanWater bool      `json:"can_water"`
	CanSteal bool      `json:"can_steal"`
	Tried    string    `json:"tried"`
}

// GardenStealRules are the limits one try is checked against.
type GardenStealRules struct {
	ProtectFor     time.Duration
	StealShare     float64
	MaxStolenShare float64
	DailyLimit     int
	// DogCatches is the dice already rolled for the owner's dog; it only counts when the dog was fed today.
	DogCatches bool
}

// GardenStealResult tells the thief how the try went.
type GardenStealResult struct {
	Caught bool    `json:"caught"`
	Amount float64 `json:"amount"`
}

// GardenSocial is the viewer's own daily state.
type GardenSocial struct {
	DogFedToday bool `json:"dog_fed_today"`
	StealsToday int  `json:"steals_today"`
	Unread      int  `json:"unread"`
	// DogCatchPercent is how likely the user's fed dog catches a thief.
	DogCatchPercent float64 `json:"dog_catch_percent"`
	// TotalRecharged is only used to work out DogCatchPercent.
	TotalRecharged float64 `json:"-"`
}

// GardenFeedItem is one line of the activity feed.
type GardenFeedItem struct {
	Kind   string  `json:"kind"`
	Amount float64 `json:"amount"`
	// Mine: the viewer did it to someone else's crop; otherwise someone did it to the viewer's crop.
	Mine bool `json:"mine"`
	// UserID is the other gardener, so the feed can lead to their garden.
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"-"`
	Email     string    `json:"-"`
}

// GardenRank is one line of this week's boards: what a gardener took, or how many thieves their dog caught.
type GardenRank struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
	// Me marks the viewer's own line.
	Me       bool   `json:"me"`
	UserID   int64  `json:"-"`
	Username string `json:"-"`
	Email    string `json:"-"`
}

// GardenRanks are this week's boards.
type GardenRanks struct {
	Thieves []GardenRank `json:"thieves"`
	Dogs    []GardenRank `json:"dogs"`
}

const gardenRankSize = 10

// GardenSocialRepository stores what neighbours do in each other's gardens.
type GardenSocialRepository interface {
	// Neighbours lists the viewer's pinned neighbours and up to perDay others who have a growing crop today.
	Neighbours(ctx context.Context, viewerID int64, perDay int) ([]GardenNeighbour, error)
	NeighbourCrops(ctx context.Context, ownerID, viewerID int64) ([]GardenNeighbourCropState, error)
	// CropOwner reports who owns a growing crop and that owner's cumulative top-ups; ok is false when there is none.
	CropOwner(ctx context.Context, cropID int64) (ownerID int64, totalRecharged float64, ok bool, err error)
	Steal(ctx context.Context, actorID, cropID int64, rules GardenStealRules) (GardenStealResult, error)
	Water(ctx context.Context, actorID, cropID int64, sooner time.Duration, maxWaters int) (time.Time, error)
	FeedDog(ctx context.Context, userID int64) error
	Social(ctx context.Context, userID int64) (GardenSocial, error)
	Feed(ctx context.Context, userID int64, limit int) ([]GardenFeedItem, error)
	// Ranks lists this week's gardeners who took the most and whose dogs caught the most.
	Ranks(ctx context.Context, limit int) (thieves, dogs []GardenRank, err error)
	// MarkFeedSeen counts what neighbours did in the user's garden up to and including until as read.
	MarkFeedSeen(ctx context.Context, userID int64, until time.Time) error
}

// gardenName is how a user appears to other gardeners: the username, or the email with its local part masked.
func gardenName(username, email string) string {
	if name := strings.TrimSpace(username); name != "" {
		return name
	}
	local, domain, found := strings.Cut(email, "@")
	if !found {
		return "***"
	}
	shown := []rune(local)
	if len(shown) > 2 {
		shown = shown[:2]
	}
	return string(shown) + "***@" + domain
}

func (s *GardenService) open(ctx context.Context) (GardenSettings, error) {
	settings := s.Settings(ctx)
	if !settings.Enabled || s.social == nil {
		return settings, ErrGardenClosed
	}
	return settings, nil
}

// Neighbours lists whose gardens the user may visit today.
func (s *GardenService) Neighbours(ctx context.Context, userID int64) ([]GardenNeighbour, error) {
	settings, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	neighbours, err := s.social.Neighbours(ctx, userID, settings.NeighboursPerDay)
	if err != nil {
		return nil, err
	}
	for i := range neighbours {
		neighbours[i].Name = gardenName(neighbours[i].Username, neighbours[i].Email)
	}
	return neighbours, nil
}

func (s *GardenService) neighbour(ctx context.Context, userID, neighbourID int64) (*GardenNeighbour, GardenSettings, error) {
	settings, err := s.open(ctx)
	if err != nil {
		return nil, settings, err
	}
	neighbours, err := s.Neighbours(ctx, userID)
	if err != nil {
		return nil, settings, err
	}
	for i := range neighbours {
		if neighbours[i].ID == neighbourID {
			return &neighbours[i], settings, nil
		}
	}
	return nil, settings, ErrGardenNotNeighbour
}

// NeighbourGarden shows one of today's neighbours' growing crops, without their amounts.
func (s *GardenService) NeighbourGarden(ctx context.Context, userID, neighbourID int64) (*GardenNeighbour, []GardenNeighbourCrop, error) {
	neighbour, settings, err := s.neighbour(ctx, userID, neighbourID)
	if err != nil {
		return nil, nil, err
	}
	states, err := s.social.NeighbourCrops(ctx, neighbourID, userID)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	protect := time.Duration(settings.ProtectHours * float64(time.Hour))
	crops := make([]GardenNeighbourCrop, 0, len(states))
	for _, state := range states {
		size := "small"
		switch {
		case state.Amount >= gardenLargeCrop:
			size = "large"
		case state.Amount >= gardenMediumCrop:
			size = "medium"
		}
		openAt := state.RipeAt.Add(protect)
		left := state.Amount*settings.MaxStolenPercent/100 - state.StolenAmount
		crops = append(crops, GardenNeighbourCrop{
			ID: state.ID, Size: size, PlantedAt: state.PlantedAt, RipeAt: state.RipeAt, OpenAt: openAt, Tried: state.Tried,
			CanWater: state.RipeAt.After(now) && !state.WateredToday && state.Waters < settings.MaxWatersPerCrop,
			CanSteal: !openAt.After(now) && state.Tried == "" && left > 0 && settings.StealPercent > 0,
		})
	}
	return neighbour, crops, nil
}

// dogChance is how likely the owner's fed dog catches a thief: it rises with the owner's growth level.
func (s *GardenService) dogChance(ctx context.Context, settings GardenSettings, ownerTotalRecharged float64) float64 {
	low, high := settings.DogMinPercent/100, settings.DogMaxPercent/100
	if s.growthLevels == nil {
		return low
	}
	ladder := s.growthLevels.Get(ctx)
	if !ladder.Enabled || len(ladder.Levels) < 2 {
		return low
	}
	level := ladder.LevelFor(ownerTotalRecharged)
	return low + (high-low)*float64(level)/float64(len(ladder.Levels)-1)
}

// Steal takes a share of a neighbour's ripe, unprotected crop into the thief's balance, unless the dog catches them.
func (s *GardenService) Steal(ctx context.Context, userID, cropID int64) (GardenStealResult, error) {
	settings, err := s.open(ctx)
	if err != nil {
		return GardenStealResult{}, err
	}
	ownerID, ownerTotal, ok, err := s.social.CropOwner(ctx, cropID)
	if err != nil {
		return GardenStealResult{}, err
	}
	if !ok {
		return GardenStealResult{}, ErrGardenCropNotFound
	}
	if ownerID == userID {
		return GardenStealResult{}, ErrGardenOwnCrop
	}
	if _, _, err := s.neighbour(ctx, userID, ownerID); err != nil {
		return GardenStealResult{}, err
	}
	result, err := s.social.Steal(ctx, userID, cropID, GardenStealRules{
		ProtectFor:     time.Duration(settings.ProtectHours * float64(time.Hour)),
		StealShare:     settings.StealPercent / 100,
		MaxStolenShare: settings.MaxStolenPercent / 100,
		DailyLimit:     settings.DailyStealLimit,
		DogCatches:     s.roll() < s.dogChance(ctx, settings, ownerTotal),
	})
	if err != nil {
		return GardenStealResult{}, err
	}
	if !result.Caught {
		s.invalidateBalance(ctx, userID)
	}
	return result, nil
}

// Water ripens a neighbour's crop sooner and reports its new ripening time.
func (s *GardenService) Water(ctx context.Context, userID, cropID int64) (time.Time, error) {
	settings, err := s.open(ctx)
	if err != nil {
		return time.Time{}, err
	}
	ownerID, _, ok, err := s.social.CropOwner(ctx, cropID)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return time.Time{}, ErrGardenCropNotFound
	}
	if ownerID == userID {
		return time.Time{}, ErrGardenOwnCrop
	}
	if _, _, err := s.neighbour(ctx, userID, ownerID); err != nil {
		return time.Time{}, err
	}
	return s.social.Water(ctx, userID, cropID, time.Duration(settings.WaterMinutes*float64(time.Minute)), settings.MaxWatersPerCrop)
}

// FeedDog feeds the user's dog for today; an unfed dog does not guard the garden.
func (s *GardenService) FeedDog(ctx context.Context, userID int64) error {
	if _, err := s.open(ctx); err != nil {
		return err
	}
	return s.social.FeedDog(ctx, userID)
}

// Social reports the user's own daily state.
func (s *GardenService) Social(ctx context.Context, userID int64) (GardenSocial, error) {
	if s.social == nil {
		return GardenSocial{}, nil
	}
	social, err := s.social.Social(ctx, userID)
	if err != nil {
		return GardenSocial{}, err
	}
	social.DogCatchPercent = math.Round(s.dogChance(ctx, s.Settings(ctx), social.TotalRecharged)*1000) / 10
	return social, nil
}

// Feed lists what recently happened in the user's garden and what the user did in others', and marks it read.
func (s *GardenService) Feed(ctx context.Context, userID int64) ([]GardenFeedItem, error) {
	if s.social == nil {
		return nil, nil
	}
	items, err := s.social.Feed(ctx, userID, gardenFeedShown)
	if err != nil {
		return nil, err
	}
	// Only what is in this list is read: something a neighbour does while it loads stays unread.
	var seenUntil time.Time
	for i := range items {
		items[i].Name = gardenName(items[i].Username, items[i].Email)
		if !items[i].Mine && items[i].CreatedAt.After(seenUntil) {
			seenUntil = items[i].CreatedAt
		}
	}
	if !seenUntil.IsZero() {
		if err := s.social.MarkFeedSeen(ctx, userID, seenUntil); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *GardenService) roll() float64 {
	if s.rollDice != nil {
		return s.rollDice()
	}
	return rand.Float64()
}

// Ranks returns this week's boards with names as strangers see them.
func (s *GardenService) Ranks(ctx context.Context, userID int64) (GardenRanks, error) {
	if _, err := s.open(ctx); err != nil {
		return GardenRanks{}, err
	}
	thieves, dogs, err := s.social.Ranks(ctx, gardenRankSize)
	if err != nil {
		return GardenRanks{}, err
	}
	for _, board := range [][]GardenRank{thieves, dogs} {
		for i := range board {
			board[i].Name = gardenName(board[i].Username, board[i].Email)
			board[i].Me = board[i].UserID == userID
		}
	}
	return GardenRanks{Thieves: thieves, Dogs: dogs}, nil
}
