package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketHandler handles the signed-in user's own support tickets.
type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Category string `json:"category"`
	Content  string `json:"content" binding:"required"`
}

type TicketReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// ticketRequestIDs resolves the caller and the :id path parameter, writing the error response itself.
func ticketRequestIDs(c *gin.Context) (userID, ticketID int64, ok bool) {
	subject, found := middleware2.GetAuthSubjectFromContext(c)
	if !found {
		response.Unauthorized(c, "User not found in context")
		return 0, 0, false
	}
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return 0, 0, false
	}
	return subject.UserID, ticketID, true
}

// List handles GET /api/v1/tickets
func (h *TicketHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.ticketService.ListForUser(c.Request.Context(), subject.UserID,
		pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.Ticket, 0, len(items))
	for i := range items {
		out = append(out, dto.TicketFromService(&items[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// Create handles POST /api/v1/tickets
func (h *TicketHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	ticket, err := h.ticketService.Create(c.Request.Context(), subject.UserID, service.CreateTicketInput{
		Subject:  req.Subject,
		Category: req.Category,
		Content:  req.Content,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.TicketFromService(ticket))
}

// Get handles GET /api/v1/tickets/:id
func (h *TicketHandler) Get(c *gin.Context) {
	userID, ticketID, ok := ticketRequestIDs(c)
	if !ok {
		return
	}
	ticket, messages, err := h.ticketService.GetForUser(c.Request.Context(), userID, ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.TicketDetail{
		Ticket:   dto.TicketFromService(ticket),
		Messages: dto.TicketMessagesFromService(messages),
	})
}

// Reply handles POST /api/v1/tickets/:id/messages
func (h *TicketHandler) Reply(c *gin.Context) {
	userID, ticketID, ok := ticketRequestIDs(c)
	if !ok {
		return
	}
	var req TicketReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	message, err := h.ticketService.ReplyAsUser(c.Request.Context(), userID, ticketID, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.TicketMessageFromService(message))
}

// Close handles POST /api/v1/tickets/:id/close
func (h *TicketHandler) Close(c *gin.Context) {
	userID, ticketID, ok := ticketRequestIDs(c)
	if !ok {
		return
	}
	if err := h.ticketService.CloseByUser(c.Request.Context(), userID, ticketID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
