package handler

import (
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// GrowthHandler serves the growth-level ladder: the signed-in user's own level, and the ladder itself to admins.
type GrowthHandler struct {
	growthLevels  *service.GrowthLevelsService
	userService   *service.UserService
	apiKeyService *service.APIKeyService
}

func NewGrowthHandler(growthLevels *service.GrowthLevelsService, userService *service.UserService, apiKeyService *service.APIKeyService) *GrowthHandler {
	return &GrowthHandler{growthLevels: growthLevels, userService: userService, apiKeyService: apiKeyService}
}

type userGrowthResponse struct {
	Enabled        bool                  `json:"enabled"`
	Levels         []service.GrowthLevel `json:"levels"`
	LevelIndex     int                   `json:"level_index"`
	TotalRecharged float64               `json:"total_recharged"`
	// ManualPrice: an admin set this user's own price on an app line, which replaces the level discount there.
	ManualPrice bool `json:"manual_price"`
}

// Get handles GET /api/v1/user/growth
func (h *GrowthHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	ladder := h.growthLevels.Get(c.Request.Context())
	if !ladder.Enabled {
		response.Success(c, userGrowthResponse{Levels: []service.GrowthLevel{}, LevelIndex: -1})
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, userGrowthResponse{
		Enabled:        true,
		Levels:         ladder.Levels,
		LevelIndex:     ladder.LevelFor(user.TotalRecharged),
		TotalRecharged: user.TotalRecharged,
		ManualPrice:    h.apiKeyService.HasManualAppLinePrice(c.Request.Context(), subject.UserID),
	})
}

// AdminGet handles GET /api/v1/admin/growth-levels
func (h *GrowthHandler) AdminGet(c *gin.Context) {
	response.Success(c, h.growthLevels.Get(c.Request.Context()))
}

// AdminUpdate handles PUT /api/v1/admin/growth-levels
func (h *GrowthHandler) AdminUpdate(c *gin.Context) {
	var ladder service.GrowthLevels
	if err := c.ShouldBindJSON(&ladder); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	saved, err := h.growthLevels.Set(c.Request.Context(), ladder)
	if errors.Is(err, service.ErrGrowthLevelsInvalid) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, saved)
}
