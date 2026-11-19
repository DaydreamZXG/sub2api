package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketHandler lets staff read and answer every user's support tickets.
type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

type TicketReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

type TicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func ticketIDParam(c *gin.Context) (int64, bool) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return 0, false
	}
	return ticketID, true
}

// List handles GET /api/v1/admin/tickets
func (h *TicketHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.ticketService.List(c.Request.Context(),
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		strings.TrimSpace(c.Query("status")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.AdminTicket, 0, len(items))
	for i := range items {
		out = append(out, dto.AdminTicketFromService(&items[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// GetByID handles GET /api/v1/admin/tickets/:id
func (h *TicketHandler) GetByID(c *gin.Context) {
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	ticket, messages, err := h.ticketService.Get(c.Request.Context(), ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminTicketDetail{
		AdminTicket: dto.AdminTicketFromService(ticket),
		Messages:    dto.TicketMessagesFromService(messages),
	})
}

// Reply handles POST /api/v1/admin/tickets/:id/messages
func (h *TicketHandler) Reply(c *gin.Context) {
	subject, found := middleware2.GetAuthSubjectFromContext(c)
	if !found {
		response.Unauthorized(c, "User not found in context")
		return
	}
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	var req TicketReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	message, err := h.ticketService.ReplyAsStaff(c.Request.Context(), subject.UserID, ticketID, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.TicketMessageFromService(message))
}

// UpdateStatus handles PUT /api/v1/admin/tickets/:id/status
func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	ticketID, ok := ticketIDParam(c)
	if !ok {
		return
	}
	var req TicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.ticketService.SetStatus(c.Request.Context(), ticketID, req.Status); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
