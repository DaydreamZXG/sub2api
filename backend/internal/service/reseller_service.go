package service

import (
	"context"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrResellerOnly       = infraerrors.Forbidden("RESELLER_ONLY", "this account is not a reseller")
	ErrResellerKeyInvalid = infraerrors.BadRequest("RESELLER_KEY_INVALID", "a client key needs a name, a budget above 0 and a markup from 1 to 100")
	ErrResellerUserGone   = infraerrors.NotFound("RESELLER_USER_NOT_FOUND", "user not found")
)

const (
	resellerMinMarkup = 1.0
	resellerMaxMarkup = 100.0
)

// ResellerEntry is one reseller as an admin sees them. Amounts are balance units.
type ResellerEntry struct {
	UserID  int64     `json:"user_id"`
	Email   string    `json:"email"`
	Since   time.Time `json:"since"`
	Balance float64   `json:"balance"`
	Keys    int       `json:"keys"`
	// ClientUsed is what their clients used at the clients' price; Cost is what that cost the reseller.
	ClientUsed float64 `json:"client_used"`
	Cost       float64 `json:"cost"`
}

// ResellerClientKey is a key a reseller made for a client. Budget and ClientUsed are in the client's
// price; Cost is what the same usage took from the reseller's balance, Margin the difference.
type ResellerClientKey struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	GroupID    *int64     `json:"group_id"`
	GroupName  string     `json:"group_name"`
	Status     string     `json:"status"`
	Markup     float64    `json:"markup"`
	Budget     float64    `json:"budget"`
	ClientUsed float64    `json:"client_used"`
	Cost       float64    `json:"cost"`
	Margin     float64    `json:"margin"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ResellerOverview is what a reseller sees on their own page.
type ResellerOverview struct {
	Balance    float64 `json:"balance"`
	ClientUsed float64 `json:"client_used"`
	Cost       float64 `json:"cost"`
	Margin     float64 `json:"margin"`
	// Reserved is what the unused part of the active keys' budgets would cost the reseller. Calls
	// are paid as they happen, so with a balance below this some clients can run dry before their budget.
	Reserved float64             `json:"reserved"`
	Keys     []ResellerClientKey `json:"keys"`
}

// ResellerClientKeyRequest is what a reseller fills in for a new client key.
type ResellerClientKeyRequest struct {
	Name          string  `json:"name"`
	GroupID       *int64  `json:"group_id"`
	Budget        float64 `json:"budget"`
	Markup        float64 `json:"markup"`
	ExpiresInDays *int    `json:"expires_in_days"`
}

type ResellerRepository interface {
	IsReseller(ctx context.Context, userID int64) (bool, error)
	// SetReseller reports false when there is no such user.
	SetReseller(ctx context.Context, userID int64, on bool) (bool, error)
	ListResellers(ctx context.Context) ([]ResellerEntry, error)
	// ClientKeys returns the user's client keys, newest first, and the user's balance.
	ClientKeys(ctx context.Context, userID int64) ([]ResellerClientKey, float64, error)
}

// resellerKeyCreator is the part of APIKeyService a reseller needs.
type resellerKeyCreator interface {
	Create(ctx context.Context, userID int64, req CreateAPIKeyRequest) (*APIKey, error)
}

// ResellerService lets chosen users hand out keys to their own clients at their own price. A call on
// such a key is paid from the reseller's balance like any other; the key's budget is used up at the
// reseller's cost times the key's markup, which is fixed when the key is made.
type ResellerService struct {
	repo ResellerRepository
	keys resellerKeyCreator
}

func NewResellerService(repo ResellerRepository, keys *APIKeyService) *ResellerService {
	return &ResellerService{repo: repo, keys: keys}
}

func (s *ResellerService) List(ctx context.Context) ([]ResellerEntry, error) {
	out, err := s.repo.ListResellers(ctx)
	if out == nil {
		out = []ResellerEntry{}
	}
	return out, err
}

// Set makes the user a reseller or stops them being one. Keys they already made keep working.
func (s *ResellerService) Set(ctx context.Context, userID int64, on bool) error {
	if userID <= 0 {
		return ErrResellerUserGone
	}
	found, err := s.repo.SetReseller(ctx, userID, on)
	if err != nil {
		return err
	}
	if !found {
		return ErrResellerUserGone
	}
	return nil
}

// Overview returns nil for a user who is not a reseller.
func (s *ResellerService) Overview(ctx context.Context, userID int64) (*ResellerOverview, error) {
	ok, err := s.repo.IsReseller(ctx, userID)
	if err != nil || !ok {
		return nil, err
	}
	keys, balance, err := s.repo.ClientKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := &ResellerOverview{Balance: balance, Keys: []ResellerClientKey{}}
	for _, k := range keys {
		k.Cost, k.Margin = resellerSplit(k.ClientUsed, k.Markup)
		out.ClientUsed += k.ClientUsed
		out.Cost += k.Cost
		out.Margin += k.Margin
		if k.Status == StatusActive && k.Budget > k.ClientUsed && (k.ExpiresAt == nil || k.ExpiresAt.After(time.Now())) {
			left, _ := resellerSplit(k.Budget-k.ClientUsed, k.Markup)
			out.Reserved += left
		}
		out.Keys = append(out.Keys, k)
	}
	return out, nil
}

// resellerSplit turns an amount in the client's price into what it cost the reseller and what is left.
func resellerSplit(clientAmount, markup float64) (cost, margin float64) {
	if markup <= 0 {
		return clientAmount, 0
	}
	cost = clientAmount / markup
	return cost, clientAmount - cost
}

// CreateClientKey makes a key whose budget is in the client's price. The markup cannot be changed later.
func (s *ResellerService) CreateClientKey(ctx context.Context, userID int64, req ResellerClientKeyRequest) (*APIKey, error) {
	ok, err := s.repo.IsReseller(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrResellerOnly
	}
	markup := math.Round(req.Markup*1e4) / 1e4
	if strings.TrimSpace(req.Name) == "" || math.IsNaN(req.Budget) || math.IsInf(req.Budget, 0) || req.Budget <= 0 ||
		math.IsNaN(markup) || markup < resellerMinMarkup || markup > resellerMaxMarkup {
		return nil, ErrResellerKeyInvalid
	}
	return s.keys.Create(ctx, userID, CreateAPIKeyRequest{
		Name:          strings.TrimSpace(req.Name),
		GroupID:       req.GroupID,
		Quota:         req.Budget,
		ExpiresInDays: req.ExpiresInDays,
		ClientMarkup:  markup,
	})
}
