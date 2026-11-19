package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Shared middleware answers in this project's own shapes; the desktop routes must still
// speak the connect protocol, which the desktop app parses strictly.
func TestDesktopProtocolErrorsRewritesForeignFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		respond    func(c *gin.Context)
		wantStatus int
		wantBody   string
	}{
		{
			name: "expired sign-in from the JWT middleware",
			respond: func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "TOKEN_EXPIRED", "message": "token has expired"})
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"success":false,"code":"session_required"}`,
		},
		{
			name: "rate limit",
			respond: func(c *gin.Context) {
				c.Header("Retry-After", "7")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded", "message": "Too many requests"})
			},
			wantStatus: http.StatusTooManyRequests,
			wantBody:   `{"success":false,"code":"rate_limited"}`,
		},
		{
			name: "other refusal keeps its reason",
			respond: func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "USER_INACTIVE", "message": "disabled"})
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `{"success":false,"code":"user_inactive"}`,
		},
		{
			name:       "protocol error passes through",
			respond:    func(c *gin.Context) { desktopError(c, http.StatusConflict, "connect_code_already_used") },
			wantStatus: http.StatusConflict,
			wantBody:   `{"success":false,"code":"connect_code_already_used"}`,
		},
		{
			name:       "success passes through",
			respond:    func(c *gin.Context) { desktopOK(c, gin.H{"code": "abc"}) },
			wantStatus: http.StatusOK,
			wantBody:   `{"success":true,"data":{"code":"abc"}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/x", (&DesktopHandler{}).ProtocolErrors, tc.respond)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

			require.Equal(t, tc.wantStatus, rec.Code)
			require.JSONEq(t, tc.wantBody, rec.Body.String())
		})
	}
}

func TestDesktopProtocolErrorsKeepsRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/x", (&DesktopHandler{}).ProtocolErrors, func(c *gin.Context) {
		c.Header("Retry-After", "7")
		c.AbortWithStatus(http.StatusTooManyRequests)
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

	require.Equal(t, "7", rec.Header().Get("Retry-After"))
	require.JSONEq(t, `{"success":false,"code":"rate_limited"}`, rec.Body.String())
}
