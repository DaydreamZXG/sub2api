//go:build unit

package service

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type agentRepoStub struct {
	AgentRepository
	accrued  int
	set      []float64
	isAgent  []bool
	created  []float64
	payout   *AgentPayout
	finished []string
	finishOK bool
	attached [][2]int64
}

func (s *agentRepoStub) AccrueShares(context.Context, time.Time) (int, error) {
	s.accrued++
	return 0, nil
}

func (s *agentRepoStub) SetAgent(_ context.Context, _ int64, isAgent bool, percent float64) error {
	s.isAgent = append(s.isAgent, isAgent)
	s.set = append(s.set, percent)
	return nil
}

func (s *agentRepoStub) CreatePayout(_ context.Context, _ int64, amount float64, _ string) (int64, error) {
	s.created = append(s.created, amount)
	return 9, nil
}

func (s *agentRepoStub) GetPayout(context.Context, int64) (*AgentPayout, error) {
	if s.payout == nil {
		return nil, ErrAgentPayoutNotFound
	}
	return s.payout, nil
}

// PayPayout as the store does it: the status only changes when withdraw went through.
func (s *agentRepoStub) PayPayout(ctx context.Context, _ int64, _ string, withdraw func(context.Context) error) (bool, error) {
	if !s.finishOK {
		return false, nil
	}
	if err := withdraw(ctx); err != nil {
		return false, err
	}
	s.finished = append(s.finished, AgentPayoutPaid)
	return true, nil
}

func (s *agentRepoStub) AgentUsers(context.Context, int64) ([]AgentInvitee, error) {
	return nil, nil
}

func (s *agentRepoStub) AttachUser(_ context.Context, agentID, userID int64) error {
	s.attached = append(s.attached, [2]int64{agentID, userID})
	return nil
}

func (s *agentRepoStub) DetachUser(context.Context, int64, int64) (bool, error) {
	return s.finishOK, nil
}

func (s *agentRepoStub) FinishPayout(_ context.Context, _ int64, status, _ string) (bool, error) {
	s.finished = append(s.finished, status)
	return s.finishOK, nil
}

type agentAffiliateRepoStub struct {
	AffiliateRepository
	withdrawals []affiliateWithdrawCall
	withdrawErr error
}

func (s *agentAffiliateRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	return &AffiliateSummary{UserID: userID}, nil
}

func (s *agentAffiliateRepoStub) WithdrawQuota(_ context.Context, userID int64, amount float64, operationID string) (*AffiliateWithdrawResult, error) {
	s.withdrawals = append(s.withdrawals, affiliateWithdrawCall{userID: userID, amount: amount, operationID: operationID})
	if s.withdrawErr != nil {
		return nil, s.withdrawErr
	}
	return &AffiliateWithdrawResult{LedgerID: 1}, nil
}

func newAgentServiceForTest(repo *agentRepoStub, aff *agentAffiliateRepoStub) *AgentService {
	return &AgentService{repo: repo, affiliate: &AffiliateService{repo: aff}, runCtx: context.Background()}
}

func TestAgentSetAgent_ChecksTheShareAndSettlesFirst(t *testing.T) {
	t.Parallel()
	for _, percent := range []float64{0, -1, 0.004, 100.01, math.NaN()} {
		repo := &agentRepoStub{}
		err := newAgentServiceForTest(repo, &agentAffiliateRepoStub{}).SetAgent(context.Background(), 7, true, percent)
		require.ErrorIs(t, err, ErrAgentShareInvalid, "percent %v", percent)
		require.Empty(t, repo.set)
	}

	repo := &agentRepoStub{}
	svc := newAgentServiceForTest(repo, &agentAffiliateRepoStub{})
	require.NoError(t, svc.SetAgent(context.Background(), 7, true, 12.345))
	// What was spent before is credited at the old terms before the new share applies.
	require.Equal(t, 1, repo.accrued)
	require.Equal(t, []float64{12.35}, repo.set)

	require.NoError(t, svc.SetAgent(context.Background(), 7, false, 55))
	require.Equal(t, []bool{true, false}, repo.isAgent)
	require.Equal(t, 0.0, repo.set[1])

	require.Error(t, svc.SetAgent(context.Background(), 0, true, 10))
}

func TestAgentRequestPayout_ChecksAmountAndNote(t *testing.T) {
	t.Parallel()
	for _, amount := range []float64{0, -5, 0.004, math.NaN(), math.Inf(1)} {
		repo := &agentRepoStub{}
		_, err := newAgentServiceForTest(repo, &agentAffiliateRepoStub{}).RequestPayout(context.Background(), 7, amount, "")
		require.ErrorIs(t, err, ErrAgentPayoutAmount, "amount %v", amount)
		require.Empty(t, repo.created)
	}

	repo := &agentRepoStub{}
	svc := newAgentServiceForTest(repo, &agentAffiliateRepoStub{})
	_, err := svc.RequestPayout(context.Background(), 7, 10, strings.Repeat("字", agentPayoutNoteRunes+1))
	require.Error(t, err)
	require.Empty(t, repo.created)

	id, err := svc.RequestPayout(context.Background(), 7, 10.128, "  支付宝 a@b.c ")
	require.NoError(t, err)
	require.Equal(t, int64(9), id)
	require.Equal(t, []float64{10.13}, repo.created)
}

func TestAgentMarkPaid_TakesTheAmountOffOnceUnderAStableKey(t *testing.T) {
	t.Parallel()
	repo := &agentRepoStub{payout: &AgentPayout{ID: 3, UserID: 7, Amount: 25, Status: AgentPayoutPending}, finishOK: true}
	aff := &agentAffiliateRepoStub{}
	svc := newAgentServiceForTest(repo, aff)

	require.NoError(t, svc.MarkPaid(context.Background(), 3, " paid "))
	require.NoError(t, svc.MarkPaid(context.Background(), 3, ""))
	require.Len(t, aff.withdrawals, 2)
	require.Equal(t, int64(7), aff.withdrawals[0].userID)
	require.Equal(t, 25.0, aff.withdrawals[0].amount)
	// Same request, same operation: the ledger takes it off only once.
	require.Equal(t, aff.withdrawals[0].operationID, aff.withdrawals[1].operationID)
	require.Equal(t, []string{AgentPayoutPaid, AgentPayoutPaid}, repo.finished)
}

func TestAgentMarkPaid_LeavesTheRequestWhenTheQuotaIsShort(t *testing.T) {
	t.Parallel()
	repo := &agentRepoStub{payout: &AgentPayout{ID: 3, UserID: 7, Amount: 25, Status: AgentPayoutPending}, finishOK: true}
	svc := newAgentServiceForTest(repo, &agentAffiliateRepoStub{withdrawErr: ErrAffiliateQuotaInsufficient})
	require.ErrorIs(t, svc.MarkPaid(context.Background(), 3, ""), ErrAffiliateQuotaInsufficient)
	require.Empty(t, repo.finished)
}

func TestAgentPayout_HandledRequestsAreNotHandledAgain(t *testing.T) {
	t.Parallel()
	aff := &agentAffiliateRepoStub{}
	repo := &agentRepoStub{payout: &AgentPayout{ID: 3, UserID: 7, Amount: 25, Status: AgentPayoutRejected}}
	svc := newAgentServiceForTest(repo, aff)
	require.ErrorIs(t, svc.MarkPaid(context.Background(), 3, ""), ErrAgentPayoutHandled)
	require.Empty(t, aff.withdrawals)
	require.ErrorIs(t, svc.Reject(context.Background(), 3, ""), ErrAgentPayoutHandled)

	missing := newAgentServiceForTest(&agentRepoStub{}, aff)
	require.ErrorIs(t, missing.MarkPaid(context.Background(), 4, ""), ErrAgentPayoutNotFound)
	require.ErrorIs(t, missing.Reject(context.Background(), 4, ""), ErrAgentPayoutNotFound)

	ok := &agentRepoStub{finishOK: true}
	require.NoError(t, newAgentServiceForTest(ok, aff).Reject(context.Background(), 3, "no"))
	require.Equal(t, []string{AgentPayoutRejected}, ok.finished)
}

// A run that never finishes on its own: it only returns when its context ends.
type agentBlockingRepoStub struct {
	agentRepoStub
	started chan struct{}
	ended   chan struct{}
}

func (s *agentBlockingRepoStub) AccrueShares(ctx context.Context, _ time.Time) (int, error) {
	close(s.started)
	<-ctx.Done()
	close(s.ended)
	return 0, ctx.Err()
}

func TestAgentServiceStopEndsARunUnderWayAndWaitsForIt(t *testing.T) {
	repo := &agentBlockingRepoStub{started: make(chan struct{}), ended: make(chan struct{})}
	svc := NewAgentService(repo, nil)
	<-repo.started

	svc.Stop()
	select {
	case <-repo.ended:
	default:
		t.Fatal("Stop returned while the run was still under way")
	}
	svc.Stop()

	var none *AgentService
	none.Stop()
}

func TestAgentServiceAttachUserSettlesFirstAndRefusesTheAgentThemselves(t *testing.T) {
	repo := &agentRepoStub{}
	svc := newAgentServiceForTest(repo, &agentAffiliateRepoStub{})

	require.ErrorIs(t, svc.AttachUser(context.Background(), 7, 7), ErrAgentUserInvalid)
	require.ErrorIs(t, svc.AttachUser(context.Background(), 7, 0), ErrAgentUserInvalid)
	require.Empty(t, repo.attached)
	require.Zero(t, repo.accrued)

	require.NoError(t, svc.AttachUser(context.Background(), 7, 9))
	// What the user's spending earned so far goes to whoever had them before the move.
	require.Equal(t, 1, repo.accrued)
	require.Equal(t, [][2]int64{{7, 9}}, repo.attached)
}

func TestAgentServiceDetachUserSaysWhenTheUserIsNotUnderTheAgent(t *testing.T) {
	missing := &agentRepoStub{}
	svc := newAgentServiceForTest(missing, &agentAffiliateRepoStub{})
	require.ErrorIs(t, svc.DetachUser(context.Background(), 7, 9), ErrAgentUserNotFound)

	found := &agentRepoStub{finishOK: true}
	svc = newAgentServiceForTest(found, &agentAffiliateRepoStub{})
	require.NoError(t, svc.DetachUser(context.Background(), 7, 9))
	require.Equal(t, 1, found.accrued)
}
