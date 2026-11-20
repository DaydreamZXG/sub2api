package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ResellerHandler serves resellers their client keys, and lets admins choose who is a reseller.
type ResellerHandler struct {
	resellers *service.ResellerService
}

func NewResellerHandler(resellers *service.ResellerService) *ResellerHandler {
	return &ResellerHandler{resellers: resellers}
}

// Get handles GET /api/v1/user/reseller. A user who is not a reseller gets {"reseller": null}.
func (h *ResellerHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	overview, err := h.resellers.Overview(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"reseller": overview})
}

// CreateKey handles POST /api/v1/user/reseller/keys: a key for a client, with a budget in the client's
// price. Stopping and deleting it go through the ordinary key endpoints.
func (h *ResellerHandler) CreateKey(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req service.ResellerClientKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	key, err := h.resellers.CreateClientKey(c.Request.Context(), subject.UserID, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": key.ID, "key": key.Key})
}

// List handles GET /api/v1/admin/resellers.
func (h *ResellerHandler) List(c *gin.Context) {
	out, err := h.resellers.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

// Add handles PUT /api/v1/admin/resellers/:user_id.
func (h *ResellerHandler) Add(c *gin.Context) {
	h.set(c, true)
}

// Remove handles DELETE /api/v1/admin/resellers/:user_id. Keys the user already made keep working.
func (h *ResellerHandler) Remove(c *gin.Context) {
	h.set(c, false)
}

func (h *ResellerHandler) set(c *gin.Context, on bool) {
	userID, ok := agentPathID(c, "user_id")
	if !ok {
		return
	}
	if err := h.resellers.Set(c.Request.Context(), userID, on); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"user_id": userID})
}
