//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type resellerRepoStub struct {
	reseller bool
	found    bool
	keys     []ResellerClientKey
	balance  float64
	set      []bool
}

func (s *resellerRepoStub) IsReseller(context.Context, int64) (bool, error) { return s.reseller, nil }
func (s *resellerRepoStub) SetReseller(_ context.Context, _ int64, on bool) (bool, error) {
	s.set = append(s.set, on)
	return s.found, nil
}
func (s *resellerRepoStub) ListResellers(context.Context) ([]ResellerEntry, error) { return nil, nil }
func (s *resellerRepoStub) ClientKeys(context.Context, int64) ([]ResellerClientKey, float64, error) {
	return s.keys, s.balance, nil
}

type resellerKeyStub struct {
	made []CreateAPIKeyRequest
}

func (s *resellerKeyStub) Create(_ context.Context, userID int64, req CreateAPIKeyRequest) (*APIKey, error) {
	s.made = append(s.made, req)
	return &APIKey{ID: 5, UserID: userID, Key: "sk-client", ClientMarkup: req.ClientMarkup}, nil
}

func TestResellerServiceOnlyResellersMakeClientKeys(t *testing.T) {
	keys := &resellerKeyStub{}
	svc := &ResellerService{repo: &resellerRepoStub{}, keys: keys}
	_, err := svc.CreateClientKey(context.Background(), 7, ResellerClientKeyRequest{Name: "a", Budget: 10, Markup: 1.5})
	require.ErrorIs(t, err, ErrResellerOnly)
	require.Empty(t, keys.made)
}

func TestResellerServiceChecksTheClientKey(t *testing.T) {
	keys := &resellerKeyStub{}
	svc := &ResellerService{repo: &resellerRepoStub{reseller: true}, keys: keys}
	for name, req := range map[string]ResellerClientKeyRequest{
		"no name":           {Name: "  ", Budget: 10, Markup: 1.5},
		"no budget":         {Name: "a", Budget: 0, Markup: 1.5},
		"below cost":        {Name: "a", Budget: 10, Markup: 0.99},
		"markup of nothing": {Name: "a", Budget: 10},
		"markup over a 100": {Name: "a", Budget: 10, Markup: 100.01},
	} {
		_, err := svc.CreateClientKey(context.Background(), 7, req)
		require.ErrorIs(t, err, ErrResellerKeyInvalid, name)
	}
	require.Empty(t, keys.made)

	group := int64(3)
	days := 30
	key, err := svc.CreateClientKey(context.Background(), 7, ResellerClientKeyRequest{Name: " shop ", GroupID: &group, Budget: 100, Markup: 1.50004, ExpiresInDays: &days})
	require.NoError(t, err)
	require.Equal(t, "sk-client", key.Key)
	require.Equal(t, []CreateAPIKeyRequest{{Name: "shop", GroupID: &group, Quota: 100, ExpiresInDays: &days, ClientMarkup: 1.5}}, keys.made)
}

func TestResellerServiceOverviewSplitsWhatClientsUsedIntoCostAndMargin(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	repo := &resellerRepoStub{reseller: true, balance: 20, keys: []ResellerClientKey{
		{ID: 1, Status: StatusActive, Markup: 1.5, Budget: 30, ClientUsed: 15},
		{ID: 2, Status: StatusActive, Markup: 2, Budget: 10, ClientUsed: 10},
		{ID: 3, Status: StatusDisabled, Markup: 2, Budget: 100, ClientUsed: 0},
		{ID: 4, Status: StatusActive, Markup: 2, Budget: 100, ClientUsed: 0, ExpiresAt: &past},
	}}
	out, err := (&ResellerService{repo: repo}).Overview(context.Background(), 7)
	require.NoError(t, err)
	require.InDelta(t, 10, out.Keys[0].Cost, 1e-9)
	require.InDelta(t, 5, out.Keys[0].Margin, 1e-9)
	require.InDelta(t, 25, out.ClientUsed, 1e-9)
	require.InDelta(t, 15, out.Cost, 1e-9)
	require.InDelta(t, 10, out.Margin, 1e-9)
	// Only the first key can still be used: 15 left at the client's price costs 10.
	require.InDelta(t, 10, out.Reserved, 1e-9)
	require.Equal(t, 20.0, out.Balance)

	none, err := (&ResellerService{repo: &resellerRepoStub{}}).Overview(context.Background(), 7)
	require.NoError(t, err)
	require.Nil(t, none)
}

func TestResellerServiceSetSaysWhenThereIsNoSuchUser(t *testing.T) {
	repo := &resellerRepoStub{}
	svc := &ResellerService{repo: repo}
	require.ErrorIs(t, svc.Set(context.Background(), 7, true), ErrResellerUserGone)
	require.ErrorIs(t, svc.Set(context.Background(), 0, true), ErrResellerUserGone)
	repo.found = true
	require.NoError(t, svc.Set(context.Background(), 7, false))
	require.Equal(t, []bool{true, false}, repo.set)
}

func TestAPIKeyChargeIsTheCostAtTheClientsPriceForAClientKey(t *testing.T) {
	cost := &CostBreakdown{ActualCost: 2}
	require.Equal(t, 2.0, (&postUsageBillingParams{Cost: cost, APIKey: &APIKey{}}).apiKeyCharge())
	require.Equal(t, 3.0, (&postUsageBillingParams{Cost: cost, APIKey: &APIKey{ClientMarkup: 1.5}}).apiKeyCharge())
}
