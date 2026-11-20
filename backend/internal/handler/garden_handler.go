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
	service.GardenSettings
	service.GardenSocial
	Now   time.Time            `json:"now"`
	Crops []service.GardenCrop `json:"crops"`
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
	social, err := h.garden.Social(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, userGardenResponse{
		GardenSettings: h.garden.Settings(c.Request.Context()),
		GardenSocial:   social,
		Now:            time.Now(),
		Crops:          crops,
	})
}

// gardenActor reads the signed-in user and the numeric :id of the request path.
func gardenActor(c *gin.Context) (userID, id int64, ok bool) {
	subject, found := middleware2.GetAuthSubjectFromContext(c)
	if !found {
		response.Unauthorized(c, "User not authenticated")
		return 0, 0, false
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return 0, 0, false
	}
	return subject.UserID, id, true
}

// Neighbours handles GET /api/v1/user/garden/neighbours
func (h *GardenHandler) Neighbours(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	neighbours, err := h.garden.Neighbours(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, neighbours)
}

// Neighbour handles GET /api/v1/user/garden/neighbours/:id
func (h *GardenHandler) Neighbour(c *gin.Context) {
	userID, neighbourID, ok := gardenActor(c)
	if !ok {
		return
	}
	neighbour, crops, err := h.garden.NeighbourGarden(c.Request.Context(), userID, neighbourID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"neighbour": neighbour, "crops": crops, "now": time.Now()})
}

// Steal handles POST /api/v1/user/garden/crops/:id/steal
func (h *GardenHandler) Steal(c *gin.Context) {
	userID, cropID, ok := gardenActor(c)
	if !ok {
		return
	}
	result, err := h.garden.Steal(c.Request.Context(), userID, cropID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// Water handles POST /api/v1/user/garden/crops/:id/water
func (h *GardenHandler) Water(c *gin.Context) {
	userID, cropID, ok := gardenActor(c)
	if !ok {
		return
	}
	ripeAt, err := h.garden.Water(c.Request.Context(), userID, cropID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ripe_at": ripeAt})
}

// FeedDog handles POST /api/v1/user/garden/dog/feed
func (h *GardenHandler) FeedDog(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if err := h.garden.FeedDog(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"dog_fed_today": true})
}

// Feed handles GET /api/v1/user/garden/feed
func (h *GardenHandler) Feed(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	items, err := h.garden.Feed(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
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
