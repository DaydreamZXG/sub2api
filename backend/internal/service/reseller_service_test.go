//go:build unit

package service

import (
	"context"
	"errors"
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

	portal      ResellerPortal
	portalErr   error
	portalReads int
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
func (s *resellerRepoStub) GetPortal(context.Context) (ResellerPortal, error) {
	s.portalReads++
	return s.portal, s.portalErr
}
func (s *resellerRepoStub) SetPortal(_ context.Context, portal ResellerPortal) error {
	s.portal = portal
	return nil
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
	line := int64(3)
	for name, req := range map[string]ResellerClientKeyRequest{
		"no name":           {Name: "  ", GroupID: &line, Budget: 10, Markup: 1.5},
		"no line":           {Name: "a", Budget: 10, Markup: 1.5},
		"no budget":         {Name: "a", GroupID: &line, Budget: 0, Markup: 1.5},
		"below cost":        {Name: "a", GroupID: &line, Budget: 10, Markup: 0.99},
		"markup of nothing": {Name: "a", GroupID: &line, Budget: 10},
		"markup over a 100": {Name: "a", GroupID: &line, Budget: 10, Markup: 100.01},
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

func TestResellerServiceSetPortalWantsABareHostAndAName(t *testing.T) {
	repo := &resellerRepoStub{}
	svc := &ResellerService{repo: repo}
	for name, portal := range map[string]ResellerPortal{
		"no name":         {Domain: "keys.example.com"},
		"not a host":      {Domain: "not a host", Name: "Keys"},
		"no dot":          {Domain: "localhost", Name: "Keys"},
		"logo not a link": {Domain: "keys.example.com", Name: "Keys", Logo: "javascript:alert(1)"},
	} {
		require.ErrorIs(t, svc.SetPortal(context.Background(), portal), ErrResellerPortalBad, name)
	}
	require.Equal(t, ResellerPortal{}, repo.portal)

	// What an admin pastes is reduced to the host.
	require.NoError(t, svc.SetPortal(context.Background(), ResellerPortal{Domain: " https://Keys.Example.com:8443/path ", Name: " Keys ", Contact: " wx: keys "}))
	require.Equal(t, ResellerPortal{Domain: "keys.example.com", Name: "Keys", Contact: "wx: keys"}, repo.portal)

	// An empty domain turns it off and needs nothing else.
	require.NoError(t, svc.SetPortal(context.Background(), ResellerPortal{}))
	require.Nil(t, portalForTest(t, svc, "keys.example.com"))
}

func TestResellerServicePortalForMatchesTheHostAndKeepsTheSettingsForAWhile(t *testing.T) {
	repo := &resellerRepoStub{portal: ResellerPortal{Domain: "keys.example.com", Name: "Keys"}}
	svc := &ResellerService{repo: repo}

	require.Nil(t, portalForTest(t, svc, "main.example.com"))
	require.Nil(t, portalForTest(t, svc, "evilkeys.example.com"))
	got := portalForTest(t, svc, "KEYS.example.com:443")
	require.NotNil(t, got)
	require.Equal(t, "Keys", got.Name)
	require.Equal(t, 1, repo.portalReads)

	// Saving through the service is seen at once, without waiting for the next reload.
	require.NoError(t, svc.SetPortal(context.Background(), ResellerPortal{Domain: "other.example.com", Name: "Other"}))
	require.Nil(t, portalForTest(t, svc, "keys.example.com"))
	require.NotNil(t, portalForTest(t, svc, "other.example.com"))
	require.Equal(t, 1, repo.portalReads)
}

func portalForTest(t *testing.T, svc *ResellerService, host string) *ResellerPortal {
	t.Helper()
	portal, err := svc.PortalFor(context.Background(), host)
	require.NoError(t, err)
	return portal
}

func TestResellerServicePortalForSaysWhenTheDomainIsNotKnownYet(t *testing.T) {
	repo := &resellerRepoStub{portal: ResellerPortal{Domain: "keys.example.com", Name: "Keys"}, portalErr: errors.New("down")}
	svc := &ResellerService{repo: repo}

	// Never read: the caller is told, and the next request asks again instead of waiting.
	_, err := svc.PortalFor(context.Background(), "keys.example.com")
	require.Error(t, err)
	_, err = svc.PortalFor(context.Background(), "keys.example.com")
	require.Error(t, err)
	require.Equal(t, 2, repo.portalReads)

	repo.portalErr = nil
	require.NotNil(t, portalForTest(t, svc, "keys.example.com"))

	// Once read, a failed reload keeps what was known.
	svc.portalLoaded = time.Now().Add(-time.Hour)
	repo.portalErr = errors.New("down")
	require.NotNil(t, portalForTest(t, svc, "keys.example.com"))
}
