package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

// An agent is an inviter who is paid a share of what the users they brought in spend, instead of the
// rebate on their top-ups. The share lands in the same affiliate quota and ledger as the rebate.

const (
	AgentPayoutPending  = "pending"
	AgentPayoutPaid     = "paid"
	AgentPayoutRejected = "rejected"

	agentAccrueEvery = 10 * time.Minute
	// Usage rows are written a little after the request they belong to.
	agentAccrueLag       = 3 * time.Minute
	agentAccrueTimeout   = 5 * time.Minute
	agentPayoutNoteRunes = 500
	agentPayoutListLimit = 200
)

var (
	ErrAgentNotAgent       = infraerrors.Forbidden("AGENT_NOT_AGENT", "this account is not an agent")
	ErrAgentShareInvalid   = infraerrors.BadRequest("AGENT_SHARE_INVALID", "share must be between 0 and 100")
	ErrAgentUserInvalid    = infraerrors.BadRequest("AGENT_USER_INVALID", "this user cannot be put under this agent")
	ErrAgentUserNotFound   = infraerrors.NotFound("AGENT_USER_NOT_FOUND", "this user is not under this agent")
	ErrAgentPayoutAmount   = infraerrors.BadRequest("AGENT_PAYOUT_AMOUNT_INVALID", "amount must be more than 0 and no more than what can be paid out")
	ErrAgentPayoutPending  = infraerrors.Conflict("AGENT_PAYOUT_PENDING", "there is already a payout request waiting")
	ErrAgentPayoutNotFound = infraerrors.NotFound("AGENT_PAYOUT_NOT_FOUND", "payout request not found")
	ErrAgentPayoutHandled  = infraerrors.Conflict("AGENT_PAYOUT_HANDLED", "this payout request was already handled")
)

// AgentEntry is one agent in the admin list. Amounts are in balance units.
type AgentEntry struct {
	UserID       int64      `json:"user_id"`
	Email        string     `json:"email"`
	SharePercent float64    `json:"share_percent"`
	Since        *time.Time `json:"since"`
	Invitees     int        `json:"invitees"`
	Shared       float64    `json:"shared"`
	Available    float64    `json:"available"`
}

// AgentInvitee is one of an agent's users with what they spent that counted and what it earned.
type AgentInvitee struct {
	UserID   int64     `json:"user_id"`
	Email    string    `json:"email"`
	JoinedAt time.Time `json:"joined_at"`
	Spent    float64   `json:"spent"`
	Share    float64   `json:"share"`
	// Attached is true when an admin put the user under the agent; JoinedAt is then when that was done.
	Attached bool `json:"attached"`
}

// AgentOverview is what an agent sees about their own business.
type AgentOverview struct {
	SharePercent float64        `json:"share_percent"`
	Since        *time.Time     `json:"since"`
	Available    float64        `json:"available"`
	Shared       float64        `json:"shared"`
	Spent        float64        `json:"spent"`
	MonthShared  float64        `json:"month_shared"`
	MonthSpent   float64        `json:"month_spent"`
	Invitees     []AgentInvitee `json:"invitees"`
	Payouts      []AgentPayout  `json:"payouts"`
}

// AgentPayout is a request to be paid out, settled by hand outside the site.
type AgentPayout struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Email     string     `json:"email"`
	Amount    float64    `json:"amount"`
	PayeeNote string     `json:"payee_note"`
	Status    string     `json:"status"`
	AdminNote string     `json:"admin_note"`
	CreatedAt time.Time  `json:"created_at"`
	HandledAt *time.Time `json:"handled_at"`
}

type AgentRepository interface {
	SetAgent(ctx context.Context, userID int64, isAgent bool, sharePercent float64) error
	ListAgents(ctx context.Context) ([]AgentEntry, error)
	// AccrueShares credits every agent their share of the usage up to upTo that was not counted yet.
	AccrueShares(ctx context.Context, upTo time.Time) (int, error)
	// Overview returns nil when the user is not an agent.
	Overview(ctx context.Context, userID int64, monthStart time.Time) (*AgentOverview, error)
	CreatePayout(ctx context.Context, userID int64, amount float64, payeeNote string) (int64, error)
	// ListPayouts lists one user's requests, or everyone's when userID is 0; status "" means any.
	ListPayouts(ctx context.Context, userID int64, status string, limit int) ([]AgentPayout, error)
	GetPayout(ctx context.Context, id int64) (*AgentPayout, error)
	// FinishPayout moves a pending request to status; false when it was not pending any more.
	FinishPayout(ctx context.Context, id int64, status, adminNote string) (bool, error)
	// AgentUsers lists the users under an agent, with what they spent and what it earned the agent.
	AgentUsers(ctx context.Context, agentID int64) ([]AgentInvitee, error)
	// AttachUser puts a user under an agent, taking them from whoever invited them, who is credited for
	// the user's spending up to that moment. The agent earns on their spending from then on.
	AttachUser(ctx context.Context, agentID, userID int64) error
	// DetachUser credits the agent for the user's spending so far and takes the user from under them;
	// false when they were not under this agent.
	DetachUser(ctx context.Context, agentID, userID int64) (bool, error)
	// PayPayout marks a pending request paid and runs withdraw in the same transaction; when withdraw
	// fails the request stays pending. False when the request was not pending any more.
	PayPayout(ctx context.Context, id int64, adminNote string, withdraw func(txCtx context.Context) error) (bool, error)
}

type AgentService struct {
	repo      AgentRepository
	affiliate *AffiliateService
	stopOnce  sync.Once
	wg        sync.WaitGroup
	// runCtx ends when the service is stopped, which also ends a run that is under way.
	runCtx context.Context
	cancel context.CancelFunc
}

// NewAgentService also starts the job that credits agents their share.
func NewAgentService(repo AgentRepository, affiliate *AffiliateService) *AgentService {
	s := &AgentService{repo: repo, affiliate: affiliate}
	s.runCtx, s.cancel = context.WithCancel(context.Background())
	if repo != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.run()
		}()
	}
	return s
}

// Stop ends the job and waits for a run under way, so nothing touches the database after it returns.
func (s *AgentService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
	s.wg.Wait()
}

func (s *AgentService) run() {
	ticker := time.NewTicker(agentAccrueEvery)
	defer ticker.Stop()
	for {
		s.accrueOnce()
		select {
		case <-s.runCtx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AgentService) accrueOnce() {
	ctx, cancel := context.WithTimeout(s.runCtx, agentAccrueTimeout)
	defer cancel()
	if _, err := s.Accrue(ctx); err != nil && s.runCtx.Err() == nil {
		logger.LegacyPrintf("service.agent", "[Agent] crediting shares failed, the next run takes it up again: %v", err)
	}
}

// Accrue credits what is due now. Runs on a timer; the admin page can also ask for it.
func (s *AgentService) Accrue(ctx context.Context) (int, error) {
	return s.repo.AccrueShares(ctx, time.Now().Add(-agentAccrueLag))
}

func (s *AgentService) SetAgent(ctx context.Context, userID int64, isAgent bool, sharePercent float64) error {
	if userID <= 0 {
		return infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	if !isAgent {
		sharePercent = 0
	} else {
		// Checked as it is stored: a share that rounds to nothing would leave the agent with no reward at all.
		if math.IsNaN(sharePercent) {
			return ErrAgentShareInvalid
		}
		if sharePercent = roundTo(sharePercent, 2); sharePercent <= 0 || sharePercent > 100 {
			return ErrAgentShareInvalid
		}
	}
	// The agent's figures live on their affiliate profile.
	if _, err := s.affiliate.EnsureUserAffiliate(ctx, userID); err != nil {
		return err
	}
	if isAgent {
		// What was spent so far belongs to the time before, at the old share or none.
		if _, err := s.Accrue(ctx); err != nil {
			return err
		}
	}
	return s.repo.SetAgent(ctx, userID, isAgent, sharePercent)
}

func (s *AgentService) ListAgents(ctx context.Context) ([]AgentEntry, error) {
	return s.repo.ListAgents(ctx)
}

// AgentUsers is the admin's view of the users under an agent, with their full e-mail addresses.
func (s *AgentService) AgentUsers(ctx context.Context, agentID int64) ([]AgentInvitee, error) {
	return s.repo.AgentUsers(ctx, agentID)
}

// AttachUser puts an existing user under an agent: for a customer the agent brought in without the
// invite link. Their spending earns the agent a share from now on, not for what they spent before.
func (s *AgentService) AttachUser(ctx context.Context, agentID, userID int64) error {
	if agentID <= 0 || userID <= 0 || agentID == userID {
		return ErrAgentUserInvalid
	}
	// The user may never have opened the invite page, so their row is made here.
	if _, err := s.affiliate.EnsureUserAffiliate(ctx, userID); err != nil {
		return err
	}
	return s.repo.AttachUser(ctx, agentID, userID)
}

// DetachUser takes a user from under an agent. What the agent was credited stays.
func (s *AgentService) DetachUser(ctx context.Context, agentID, userID int64) error {
	done, err := s.repo.DetachUser(ctx, agentID, userID)
	if err != nil {
		return err
	}
	if !done {
		return ErrAgentUserNotFound
	}
	return nil
}

// Overview returns nil for a user who is not an agent.
func (s *AgentService) Overview(ctx context.Context, userID int64) (*AgentOverview, error) {
	now := timezone.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	out, err := s.repo.Overview(ctx, userID, monthStart)
	if err != nil || out == nil {
		return nil, err
	}
	for i := range out.Invitees {
		out.Invitees[i].Email = maskEmail(out.Invitees[i].Email)
	}
	if out.Payouts, err = s.repo.ListPayouts(ctx, userID, "", agentPayoutListLimit); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AgentService) RequestPayout(ctx context.Context, userID int64, amount float64, payeeNote string) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, ErrAgentPayoutAmount
	}
	amount = roundTo(amount, 2)
	if amount <= 0 {
		return 0, ErrAgentPayoutAmount
	}
	payeeNote = strings.TrimSpace(payeeNote)
	if len([]rune(payeeNote)) > agentPayoutNoteRunes {
		return 0, infraerrors.BadRequest("AGENT_PAYOUT_NOTE_TOO_LONG", "the note is too long")
	}
	return s.repo.CreatePayout(ctx, userID, amount, payeeNote)
}

func (s *AgentService) ListPayouts(ctx context.Context, status string) ([]AgentPayout, error) {
	switch status {
	case "", AgentPayoutPending, AgentPayoutPaid, AgentPayoutRejected:
	default:
		return nil, infraerrors.BadRequest("AGENT_PAYOUT_STATUS_INVALID", "unknown status")
	}
	return s.repo.ListPayouts(ctx, 0, status, agentPayoutListLimit)
}

// MarkPaid records that a request was paid outside the site: the amount leaves the agent's quota.
func (s *AgentService) MarkPaid(ctx context.Context, id int64, adminNote string) error {
	payout, err := s.repo.GetPayout(ctx, id)
	if err != nil {
		return err
	}
	if payout.Status != AgentPayoutPending {
		return ErrAgentPayoutHandled
	}
	// One transaction for both: the request becomes paid and the amount leaves the quota, or neither.
	// A reject arriving at the same moment waits for it and then finds the request handled.
	done, err := s.repo.PayPayout(ctx, id, strings.TrimSpace(adminNote), func(txCtx context.Context) error {
		_, err := s.affiliate.AdminWithdrawQuota(txCtx, payout.UserID, payout.Amount, fmt.Sprintf("agent-payout-%d", id))
		return err
	})
	if err != nil {
		return err
	}
	if !done {
		return ErrAgentPayoutHandled
	}
	return nil
}

func (s *AgentService) Reject(ctx context.Context, id int64, adminNote string) error {
	done, err := s.repo.FinishPayout(ctx, id, AgentPayoutRejected, strings.TrimSpace(adminNote))
	if err != nil {
		return err
	}
	if !done {
		if _, err := s.repo.GetPayout(ctx, id); err != nil {
			return err
		}
		return ErrAgentPayoutHandled
	}
	return nil
}
