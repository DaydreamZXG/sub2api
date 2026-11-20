package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// GardenHandler serves the garden: the signed-in user's crops, and the rules to admins.
type GardenHandler struct {
	garden *service.GardenService
}

func NewGardenHandler(garden *service.GardenService) *GardenHandler {
	return &GardenHandler{garden: garden}
}

type userGardenResponse struct {
	Enabled    bool                 `json:"enabled"`
	RipenHours float64              `json:"ripen_hours"`
	Now        time.Time            `json:"now"`
	Crops      []service.GardenCrop `json:"crops"`
}

// Get handles GET /api/v1/user/garden
func (h *GardenHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	crops, err := h.garden.Crops(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	settings := h.garden.Settings(c.Request.Context())
	response.Success(c, userGardenResponse{
		Enabled:    settings.Enabled,
		RipenHours: settings.RipenHours,
		Now:        time.Now(),
		Crops:      crops,
	})
}

// Harvest handles POST /api/v1/user/garden/crops/:id/harvest
func (h *GardenHandler) Harvest(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	cropID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || cropID <= 0 {
		response.BadRequest(c, "Invalid crop id")
		return
	}
	amount, err := h.garden.Harvest(c.Request.Context(), subject.UserID, cropID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"amount": amount})
}

// AdminGet handles GET /api/v1/admin/garden-settings
func (h *GardenHandler) AdminGet(c *gin.Context) {
	response.Success(c, h.garden.Settings(c.Request.Context()))
}

// AdminUpdate handles PUT /api/v1/admin/garden-settings
func (h *GardenHandler) AdminUpdate(c *gin.Context) {
	var settings service.GardenSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	saved, err := h.garden.SetSettings(c.Request.Context(), settings)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, saved)
}
