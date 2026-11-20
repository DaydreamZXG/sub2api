package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ProfitHandler lets an admin record what upstream capacity cost and read the profit report.
type ProfitHandler struct {
	profitService *service.ProfitService
}

func NewProfitHandler(profitService *service.ProfitService) *ProfitHandler {
	return &ProfitHandler{profitService: profitService}
}

type costPackPaymentRequest struct {
	Amount float64 `json:"amount"`
	PaidAt string  `json:"paid_at"`
	Note   string  `json:"note"`
}

func profitPathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ID")
		return 0, false
	}
	return id, true
}

// Report handles GET /admin/profit/report?from=2006-01-02&to=2006-01-02.
func (h *ProfitHandler) Report(c *gin.Context) {
	report, err := h.profitService.Report(c.Request.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, report)
}

// Rebuild handles POST /admin/profit/rebuild: sum the usage logs still kept into the report again.
func (h *ProfitHandler) Rebuild(c *gin.Context) {
	if err := h.profitService.Rebuild(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"rebuilt": true})
}

func (h *ProfitHandler) GetSettings(c *gin.Context) {
	response.Success(c, h.profitService.GetSettings(c.Request.Context()))
}

func (h *ProfitHandler) UpdateSettings(c *gin.Context) {
	var req service.ProfitSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.profitService.UpdateSettings(c.Request.Context(), req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, req)
}

func (h *ProfitHandler) ListPacks(c *gin.Context) {
	packs, err := h.profitService.ListPacks(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, packs)
}

func (h *ProfitHandler) CreatePack(c *gin.Context) {
	var req service.CostPackInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	id, err := h.profitService.CreatePack(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{"id": id})
}

func (h *ProfitHandler) UpdatePack(c *gin.Context) {
	id, ok := profitPathID(c, "id")
	if !ok {
		return
	}
	var req service.CostPackInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.profitService.UpdatePack(c.Request.Context(), id, req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *ProfitHandler) DeletePack(c *gin.Context) {
	id, ok := profitPathID(c, "id")
	if !ok {
		return
	}
	if err := h.profitService.DeletePack(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *ProfitHandler) AddPayment(c *gin.Context) {
	id, ok := profitPathID(c, "id")
	if !ok {
		return
	}
	var req costPackPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.profitService.AddPayment(c.Request.Context(), id, req.Amount, req.PaidAt, req.Note); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{"id": id})
}

func (h *ProfitHandler) DeletePayment(c *gin.Context) {
	id, ok := profitPathID(c, "id")
	if !ok {
		return
	}
	paymentID, ok := profitPathID(c, "payment_id")
	if !ok {
		return
	}
	if err := h.profitService.DeletePayment(c.Request.Context(), id, paymentID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
