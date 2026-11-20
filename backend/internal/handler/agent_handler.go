package handler

import (
	"context"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AgentHandler serves agents their own figures and payout requests, and lets admins run the scheme.
type AgentHandler struct {
	agents *service.AgentService
}

func NewAgentHandler(agents *service.AgentService) *AgentHandler {
	return &AgentHandler{agents: agents}
}

type agentPayoutRequest struct {
	Amount    float64 `json:"amount"`
	PayeeNote string  `json:"payee_note"`
}

type agentSetRequest struct {
	SharePercent float64 `json:"share_percent"`
	// Tiers replaces the agent's ladder; left out or empty means one flat share.
	Tiers []service.AgentTier `json:"tiers"`
}

type agentPayoutDecision struct {
	AdminNote string `json:"admin_note"`
}

func agentPathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ID")
		return 0, false
	}
	return id, true
}

// Get handles GET /api/v1/user/agent. A user who is not an agent gets {"agent": null}.
func (h *AgentHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	overview, err := h.agents.Overview(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"agent": overview})
}

// RequestPayout handles POST /api/v1/user/agent/payouts.
func (h *AgentHandler) RequestPayout(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req agentPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	id, err := h.agents.RequestPayout(c.Request.Context(), subject.UserID, req.Amount, req.PayeeNote)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// List handles GET /api/v1/admin/agents.
func (h *AgentHandler) List(c *gin.Context) {
	agents, err := h.agents.ListAgents(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, agents)
}

// Set handles PUT /api/v1/admin/agents/:user_id: make the user an agent, or change their share.
func (h *AgentHandler) Set(c *gin.Context) {
	userID, ok := agentPathID(c, "user_id")
	if !ok {
		return
	}
	var req agentSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.agents.SetAgent(c.Request.Context(), userID, true, req.SharePercent, req.Tiers); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID})
}

// Remove handles DELETE /api/v1/admin/agents/:user_id. What was already credited stays.
func (h *AgentHandler) Remove(c *gin.Context) {
	userID, ok := agentPathID(c, "user_id")
	if !ok {
		return
	}
	if err := h.agents.SetAgent(c.Request.Context(), userID, false, 0, nil); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID})
}

// ListUsers handles GET /api/v1/admin/agents/:user_id/users.
func (h *AgentHandler) ListUsers(c *gin.Context) {
	agentID, ok := agentPathID(c, "user_id")
	if !ok {
		return
	}
	users, err := h.agents.AgentUsers(c.Request.Context(), agentID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, users)
}

// AttachUser handles PUT /api/v1/admin/agents/:user_id/users/:invitee_id: put an existing user under the agent.
func (h *AgentHandler) AttachUser(c *gin.Context) {
	h.changeUser(c, h.agents.AttachUser)
}

// DetachUser handles DELETE /api/v1/admin/agents/:user_id/users/:invitee_id.
func (h *AgentHandler) DetachUser(c *gin.Context) {
	h.changeUser(c, h.agents.DetachUser)
}

func (h *AgentHandler) changeUser(c *gin.Context, apply func(ctx context.Context, agentID, userID int64) error) {
	agentID, ok := agentPathID(c, "user_id")
	if !ok {
		return
	}
	userID, ok := agentPathID(c, "invitee_id")
	if !ok {
		return
	}
	if err := apply(c.Request.Context(), agentID, userID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID})
}

// ListPayouts handles GET /api/v1/admin/agents/payouts?status=pending.
func (h *AgentHandler) ListPayouts(c *gin.Context) {
	payouts, err := h.agents.ListPayouts(c.Request.Context(), c.Query("status"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, payouts)
}

// MarkPaid handles POST /api/v1/admin/agents/payouts/:id/paid.
func (h *AgentHandler) MarkPaid(c *gin.Context) {
	h.decide(c, h.agents.MarkPaid)
}

// Reject handles POST /api/v1/admin/agents/payouts/:id/reject.
func (h *AgentHandler) Reject(c *gin.Context) {
	h.decide(c, h.agents.Reject)
}

func (h *AgentHandler) decide(c *gin.Context, apply func(ctx context.Context, id int64, note string) error) {
	id, ok := agentPathID(c, "id")
	if !ok {
		return
	}
	var req agentPayoutDecision
	// The note is optional, so an empty body is fine.
	_ = c.ShouldBindJSON(&req)
	if err := apply(c.Request.Context(), id, req.AdminNote); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
