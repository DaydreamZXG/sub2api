package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// OfficialPriceBookPayload is the whole site-wide official price book.
type OfficialPriceBookPayload struct {
	Entries *[]service.OfficialPriceEntry `json:"entries"`
}

// GetOfficialPriceBook GET /api/v1/admin/settings/official-prices
func (h *SettingHandler) GetOfficialPriceBook(c *gin.Context) {
	entries, err := h.settingService.GetOfficialPriceBook(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"entries": entries})
}

// UpdateOfficialPriceBook replaces the whole book; an empty list clears it.
// PUT /api/v1/admin/settings/official-prices
func (h *SettingHandler) UpdateOfficialPriceBook(c *gin.Context) {
	var req OfficialPriceBookPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Entries == nil {
		response.BadRequest(c, "entries is required")
		return
	}
	saved, err := h.settingService.UpdateOfficialPriceBook(c.Request.Context(), *req.Entries)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"entries": saved})
}
