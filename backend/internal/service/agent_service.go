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
}

type AgentService struct {
	repo      AgentRepository
	affiliate *AffiliateService
	stop      chan struct{}
	stopOnce  sync.Once
}

// NewAgentService also starts the job that credits agents their share.
func NewAgentService(repo AgentRepository, affiliate *AffiliateService) *AgentService {
	s := &AgentService{repo: repo, affiliate: affiliate, stop: make(chan struct{})}
	if repo != nil {
		go s.run()
	}
	return s
}

func (s *AgentService) Stop() {
	s.stopOnce.Do(func() { close(s.stop) })
}

func (s *AgentService) run() {
	ticker := time.NewTicker(agentAccrueEvery)
	defer ticker.Stop()
	for {
		s.accrueOnce()
		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}
	}
}

func (s *AgentService) accrueOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), agentAccrueTimeout)
	defer cancel()
	if _, err := s.Accrue(ctx); err != nil {
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
	} else if math.IsNaN(sharePercent) || sharePercent <= 0 || sharePercent > 100 {
		return ErrAgentShareInvalid
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
	return s.repo.SetAgent(ctx, userID, isAgent, roundTo(sharePercent, 2))
}

func (s *AgentService) ListAgents(ctx context.Context) ([]AgentEntry, error) {
	return s.repo.ListAgents(ctx)
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
	// The key makes the deduction happen once, also when the status update below has to be repeated.
	if _, err := s.affiliate.AdminWithdrawQuota(ctx, payout.UserID, payout.Amount, fmt.Sprintf("agent-payout-%d", id)); err != nil {
		return err
	}
	done, err := s.repo.FinishPayout(ctx, id, AgentPayoutPaid, strings.TrimSpace(adminNote))
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
