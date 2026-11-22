package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/middleware"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterDesktopRoutes 注册桌面助手“一键连接”协议的站点侧接口。
// 路径由桌面助手的公开协议规定，所以不在 /api/v1 下。
func RegisterDesktopRoutes(r *gin.Engine, h *handler.Handlers, jwtAuth servermiddleware.JWTAuthMiddleware, redisClient *redis.Client) {
	rateLimiter := middleware.NewRateLimiter(redisClient)
	desktop := r.Group("/api/desktop/v2")
	desktop.Use(h.Desktop.ProtocolErrors)

	// 换取会话：由桌面助手调用，无登录态，凭一次性连接码。
	desktop.POST("/connect-codes/exchange", rateLimiter.LimitWithOptions("desktop-connect-exchange", 20, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailClose,
	}), h.Desktop.ExchangeConnectCode)

	// 站点自述：桌面助手靠它认出这个站支持授权连接，以及各接口在哪。
	desktop.GET("/bootstrap", h.Desktop.Bootstrap)
	// 续登录：凭 refresh token，无登录态。
	desktop.POST("/sessions/refresh", rateLimiter.LimitWithOptions("desktop-session-refresh", 30, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailClose,
	}), h.Desktop.RefreshSession)

	signedIn := desktop.Group("")
	signedIn.Use(gin.HandlerFunc(jwtAuth))
	signedIn.DELETE("/sessions/current", h.Desktop.EndSession)
	signedIn.GET("/account", h.Desktop.Account)
	signedIn.GET("/plans", h.Desktop.Plans)
	signedIn.GET("/keys", h.Desktop.ListKeys)
	signedIn.POST("/keys", rateLimiter.LimitWithOptions("desktop-key-create", 30, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailClose,
	}), h.Desktop.CreateKey)
	signedIn.POST("/keys/:id/reveal", h.Desktop.RevealKey)
	signedIn.DELETE("/keys/:id", h.Desktop.DeleteKey)
	signedIn.GET("/usage/summary", h.Desktop.UsageSummary)
	signedIn.GET("/usage/records", h.Desktop.UsageRecords)
	signedIn.GET("/notices", h.Desktop.Notices)
	signedIn.POST("/connect-codes", rateLimiter.LimitWithOptions("desktop-connect-issue", 20, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailClose,
	}), h.Desktop.IssueConnectCode)
	signedIn.POST("/app-tokens", rateLimiter.LimitWithOptions("desktop-app-token", 20, time.Minute, middleware.RateLimitOptions{
		FailureMode: middleware.RateLimitFailClose,
	}), h.Desktop.IssueAppToken)
}
