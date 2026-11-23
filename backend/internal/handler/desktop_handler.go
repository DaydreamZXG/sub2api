package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DesktopHandler is the station side of the desktop assistant's connect protocol.
// Its responses use the protocol's own envelope ({success, data} / {success, code}),
// not this project's usual one, because the desktop app parses them strictly.
type DesktopHandler struct {
	connectService      *service.DesktopConnectService
	authService         *service.AuthService
	userService         *service.UserService
	apiKeyService       *service.APIKeyService
	plazaService        *service.ModelPlazaService
	usageService        *service.UsageService
	announcementService *service.AnnouncementService
	settingService      *service.SettingService
}

func NewDesktopHandler(
	connectService *service.DesktopConnectService,
	authService *service.AuthService,
	userService *service.UserService,
	apiKeyService *service.APIKeyService,
	plazaService *service.ModelPlazaService,
	usageService *service.UsageService,
	announcementService *service.AnnouncementService,
	settingService *service.SettingService,
) *DesktopHandler {
	return &DesktopHandler{
		connectService:      connectService,
		authService:         authService,
		userService:         userService,
		apiKeyService:       apiKeyService,
		plazaService:        plazaService,
		usageService:        usageService,
		announcementService: announcementService,
		settingService:      settingService,
	}
}

type issueConnectCodeRequest struct {
	Origin string `json:"origin"`
}

type exchangeConnectCodeRequest struct {
	Code string `json:"code"`
}

type issueAppTokenRequest struct {
	DeviceID      string `json:"deviceId"`
	Platform      string `json:"platform"`
	ClientVersion string `json:"clientVersion"`
}

func desktopError(c *gin.Context, status int, code string) {
	c.AbortWithStatusJSON(status, gin.H{"success": false, "code": code})
}

func desktopOK(c *gin.Context, data gin.H) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// desktopResponseBuffer holds a response back so ProtocolErrors can rewrite it.
type desktopResponseBuffer struct {
	gin.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *desktopResponseBuffer) WriteHeader(status int)         { w.status = status }
func (w *desktopResponseBuffer) WriteHeaderNow()                {}
func (w *desktopResponseBuffer) Status() int                    { return w.status }
func (w *desktopResponseBuffer) Write(data []byte) (int, error) { return w.body.Write(data) }
func (w *desktopResponseBuffer) WriteString(data string) (int, error) {
	return w.body.WriteString(data)
}

// ProtocolErrors makes failures raised by shared middleware (sign-in checks, rate limits) answer in
// the protocol's envelope too; those middlewares otherwise write this project's own error shapes.
func (h *DesktopHandler) ProtocolErrors(c *gin.Context) {
	buffer := &desktopResponseBuffer{ResponseWriter: c.Writer, status: http.StatusOK}
	c.Writer = buffer
	c.Next()
	c.Writer = buffer.ResponseWriter

	body := buffer.body.Bytes()
	if buffer.status >= http.StatusBadRequest {
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if _, isProtocol := parsed["success"]; !isProtocol {
			code := "request_failed"
			switch {
			case buffer.status == http.StatusUnauthorized:
				code = "session_required"
			case buffer.status == http.StatusTooManyRequests:
				code = "rate_limited"
			case buffer.status >= http.StatusInternalServerError:
				code = "internal_error"
			default:
				if reason, ok := parsed["code"].(string); ok && reason != "" {
					code = strings.ToLower(reason)
				}
			}
			body, _ = json.Marshal(gin.H{"success": false, "code": code})
			c.Header("Content-Type", "application/json; charset=utf-8")
		}
	}
	c.Writer.WriteHeader(buffer.status)
	_, _ = c.Writer.Write(body)
}

// IssueConnectCode handles POST /api/desktop/v2/connect-codes
func (h *DesktopHandler) IssueConnectCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	var req issueConnectCodeRequest
	_ = c.ShouldBindJSON(&req)
	if req.Origin == "" {
		req.Origin = c.GetHeader("Origin")
	}
	issued, err := h.connectService.IssueCode(c.Request.Context(), subject.UserID, req.Origin)
	if errors.Is(err, service.ErrDesktopOriginInvalid) {
		desktopError(c, http.StatusBadRequest, "connect_origin_invalid")
		return
	}
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	desktopOK(c, gin.H{
		"code":       issued.Code,
		"origin":     issued.Origin,
		"expiresAt":  issued.ExpiresAt.Unix(),
		"ttlSeconds": int(service.DesktopConnectCodeTTL.Seconds()),
	})
}

// ExchangeConnectCode handles POST /api/desktop/v2/connect-codes/exchange
func (h *DesktopHandler) ExchangeConnectCode(c *gin.Context) {
	var req exchangeConnectCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Code) < 32 || len(req.Code) > 512 {
		desktopError(c, http.StatusBadRequest, "connect_code_expired")
		return
	}
	bundle, err := h.connectService.Exchange(c.Request.Context(), req.Code, c.Request.Host)
	switch {
	case errors.Is(err, service.ErrDesktopCodeAlreadyUsed):
		desktopError(c, http.StatusConflict, "connect_code_already_used")
	case errors.Is(err, service.ErrDesktopCodeExpired):
		desktopError(c, http.StatusBadRequest, "connect_code_expired")
	case errors.Is(err, service.ErrDesktopOriginMismatch):
		desktopError(c, http.StatusForbidden, "connect_origin_mismatch")
	case errors.Is(err, service.ErrUserNotActive):
		desktopError(c, http.StatusForbidden, "user_inactive")
	case err != nil:
		desktopError(c, http.StatusInternalServerError, "internal_error")
	default:
		desktopOK(c, gin.H{
			"accessToken":      bundle.AccessToken,
			"tokenType":        "Bearer",
			"accessExpiresAt":  bundle.AccessExpiresAt.Unix(),
			"refreshToken":     bundle.RefreshToken,
			"refreshExpiresAt": bundle.RefreshExpiresAt.Unix(),
			"sessionId":        bundle.SessionID,
		})
	}
}

// IssueAppToken handles POST /api/desktop/v2/app-tokens
func (h *DesktopHandler) IssueAppToken(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	var req issueAppTokenRequest
	_ = c.ShouldBindJSON(&req)
	device := service.AppDevice{DeviceID: req.DeviceID, Platform: req.Platform, ClientVersion: req.ClientVersion}
	token, err := h.connectService.IssueAppToken(c.Request.Context(), subject.UserID, device)
	if errors.Is(err, service.ErrDesktopDeviceInvalid) {
		desktopError(c, http.StatusBadRequest, "device_invalid")
		return
	}
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	desktopOK(c, gin.H{"appToken": token, "deviceId": device.DeviceID})
}
