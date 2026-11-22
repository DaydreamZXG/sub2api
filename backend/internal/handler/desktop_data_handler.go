package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// The data half of the desktop connect protocol: what the desktop app shows after sign-in.
// Every amount leaves here already converted to the currency users pay in, so the app never
// has to learn this project's balance units.

const (
	desktopService       = "desktop-connect"
	desktopContractID    = "desktop-connect-v1"
	desktopBasePath      = "/api/desktop/v2"
	desktopWalletPath    = "/topup"
	desktopUsageDays     = 30
	desktopMaxPageSize   = 100
	desktopKeyNameMaxLen = 100
	perMillionTokens     = 1_000_000
	// Used to restate official USD prices in yuan when the site has not set its own rate.
	defaultCNYPerUSD = 7.2
)

// desktopMoney converts stored amounts into what the user sees.
type desktopMoney struct {
	Currency     string
	UnitsPerYuan float64
	CNYPerUSD    float64
}

func newDesktopMoney(settings *service.PublicSettings) desktopMoney {
	if settings == nil || !settings.PaymentEnabled || settings.BalanceRechargeMultiplier <= 0 {
		return desktopMoney{Currency: "USD"}
	}
	rate := settings.SubscriptionUSDToCNYRate
	if rate <= 0 {
		rate = defaultCNYPerUSD
	}
	return desktopMoney{Currency: "CNY", UnitsPerYuan: settings.BalanceRechargeMultiplier, CNYPerUSD: rate}
}

// charged converts an amount held in balance units.
func (m desktopMoney) charged(units float64) float64 {
	if m.Currency != "CNY" {
		return roundMoney(units)
	}
	return roundMoney(units / m.UnitsPerYuan)
}

// official converts an upstream list price in USD.
func (m desktopMoney) official(usd float64) float64 {
	if m.Currency != "CNY" {
		return roundMoney(usd)
	}
	return roundMoney(usd * m.CNYPerUSD)
}

func roundMoney(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

func (h *DesktopHandler) money(c *gin.Context) desktopMoney {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		return desktopMoney{Currency: "USD"}
	}
	return newDesktopMoney(settings)
}

// Bootstrap handles GET /api/desktop/v2/bootstrap
func (h *DesktopHandler) Bootstrap(c *gin.Context) {
	origin, err := service.CanonicalDesktopOrigin("https://" + c.Request.Host)
	if err != nil {
		desktopError(c, http.StatusBadRequest, "connect_origin_invalid")
		return
	}
	desktopOK(c, gin.H{
		"schema_version":                 2,
		"service":                        desktopService,
		"contract_id":                    desktopContractID,
		"minimum_client_version":         "0.0.0",
		"device_authorization_available": false,
		"oauth_loopback_pkce_available":  false,
		"account_read":                   true,
		"usage_read":                     true,
		"models_read":                    true,
		"pricing_read":                   true,
		"tool_keys_manage":               true,
		"session_refresh_path":           desktopBasePath + "/sessions/refresh",
		"session_logout_path":            desktopBasePath + "/sessions/current",
		"account_path":                   desktopBasePath + "/account",
		"plans_path":                     desktopBasePath + "/plans",
		"keys_path":                      desktopBasePath + "/keys",
		"usage_summary_path":             desktopBasePath + "/usage/summary",
		"usage_records_path":             desktopBasePath + "/usage/records",
		"announcements_path":             desktopBasePath + "/notices",
		"app_tokens_path":                desktopBasePath + "/app-tokens",
		"app_token_header":               service.AppTokenHeader,
		"wallet_url":                     origin + desktopWalletPath,
	})
}

type refreshDesktopSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
}

// RefreshSession handles POST /api/desktop/v2/sessions/refresh
func (h *DesktopHandler) RefreshSession(c *gin.Context) {
	var req refreshDesktopSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" || req.SessionID == "" {
		desktopError(c, http.StatusBadRequest, "invalid_request")
		return
	}
	pair, err := h.authService.RefreshTokenPair(c.Request.Context(), req.RefreshToken)
	if err != nil {
		desktopError(c, http.StatusUnauthorized, "session_expired")
		return
	}
	if h.settingService.IsBackendModeEnabled(c.Request.Context()) && pair.UserRole != service.RoleAdmin {
		desktopError(c, http.StatusForbidden, "backend_mode")
		return
	}
	now := time.Now()
	desktopOK(c, gin.H{
		"access_token":       pair.AccessToken,
		"token_type":         "Bearer",
		"access_expires_at":  now.Add(time.Duration(pair.ExpiresIn) * time.Second).Unix(),
		"refresh_token":      pair.RefreshToken,
		"refresh_expires_at": now.Add(h.authService.RefreshTokenTTL()).Unix(),
		"session_id":         req.SessionID,
	})
}

// EndSession handles DELETE /api/desktop/v2/sessions/current
func (h *DesktopHandler) EndSession(c *gin.Context) {
	if sessionID := c.GetString(middleware2.ContextKeySessionID); sessionID != "" {
		if err := h.authService.RevokeSessionFamily(c.Request.Context(), sessionID); err != nil {
			desktopError(c, http.StatusInternalServerError, "internal_error")
			return
		}
	}
	desktopOK(c, gin.H{})
}

// Account handles GET /api/desktop/v2/account
func (h *DesktopHandler) Account(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	money := h.money(c)
	displayName := user.Username
	if displayName == "" {
		displayName = user.Email
	}
	desktopOK(c, gin.H{
		"username":    user.Username,
		"displayName": displayName,
		"email":       user.Email,
		"currency":    money.Currency,
		"balance":     money.charged(user.Balance),
	})
}

type desktopPrice struct {
	Input      *float64 `json:"input,omitempty"`
	Output     *float64 `json:"output,omitempty"`
	CacheRead  *float64 `json:"cacheRead,omitempty"`
	CacheWrite *float64 `json:"cacheWrite,omitempty"`
	PerRequest *float64 `json:"perRequest,omitempty"`
}

type desktopLongContextTier struct {
	FromTokens int          `json:"fromTokens"`
	ToTokens   *int         `json:"toTokens,omitempty"`
	Price      desktopPrice `json:"price"`
}

type desktopPeak struct {
	Start        string  `json:"start"`
	End          string  `json:"end"`
	Multiplier   float64 `json:"multiplier"`
	Timezone     string  `json:"timezone,omitempty"`
	WeekdaysOnly bool    `json:"weekdaysOnly,omitempty"`
}

type desktopPlanModel struct {
	ID          string                   `json:"id"`
	Price       *desktopPrice            `json:"price,omitempty"`
	Official    *desktopPrice            `json:"official,omitempty"`
	LongContext []desktopLongContextTier `json:"longContext,omitempty"`
	Peak        []desktopPeak            `json:"peak,omitempty"`
}

type desktopPlan struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Platform     string             `json:"platform"`
	Formats      []string           `json:"formats"`
	Subscription bool               `json:"subscription"`
	AppOnly      bool               `json:"appOnly"`
	AppFallback  bool               `json:"appFallback"`
	Multiplier   float64            `json:"multiplier"`
	Peak         []desktopPeak      `json:"peak,omitempty"`
	Models       []desktopPlanModel `json:"models"`
}

// desktopFormats lists the request formats a plan's platform serves natively.
// An empty list means "not stated": the app finds out by trying.
func desktopFormats(platform string) []string {
	switch platform {
	case domain.PlatformAnthropic:
		return []string{"anthropic"}
	case domain.PlatformOpenAI:
		return []string{"openai-chat", "openai-responses"}
	case domain.PlatformGemini:
		return []string{"gemini"}
	default:
		return []string{}
	}
}

func desktopPerMillion(perToken *float64, convert func(float64) float64, rate float64) *float64 {
	if perToken == nil {
		return nil
	}
	v := convert(*perToken * rate * perMillionTokens)
	return &v
}

func toDesktopPlan(g *service.PlazaGroup, rate float64, money desktopMoney) desktopPlan {
	plan := desktopPlan{
		ID:           strconv.FormatInt(g.ID, 10),
		Name:         g.Name,
		Description:  g.Description,
		Platform:     g.Platform,
		Formats:      desktopFormats(g.Platform),
		Subscription: g.SubscriptionType == service.SubscriptionTypeSubscription,
		AppOnly:      g.AppOnly,
		AppFallback:  g.AppOnly && g.AppFallback,
		Multiplier:   rate,
		Models:       make([]desktopPlanModel, 0, len(g.Models)),
	}
	if g.PeakRateEnabled && g.PeakRateMultiplier > 0 && g.PeakRateMultiplier != 1 {
		plan.Peak = []desktopPeak{{Start: g.PeakStart, End: g.PeakEnd, Multiplier: g.PeakRateMultiplier}}
	}
	for i := range g.Models {
		m := &g.Models[i]
		model := desktopPlanModel{ID: m.Name}
		if p := m.Pricing; p != nil {
			model.Price = &desktopPrice{
				Input:      desktopPerMillion(p.InputPrice, money.charged, rate),
				Output:     desktopPerMillion(p.OutputPrice, money.charged, rate),
				CacheRead:  desktopPerMillion(p.CacheReadPrice, money.charged, rate),
				CacheWrite: desktopPerMillion(p.CacheWritePrice, money.charged, rate),
			}
			if p.PerRequestPrice != nil {
				v := money.charged(*p.PerRequestPrice * rate)
				model.Price.PerRequest = &v
			}
			if g.LongContextPricingEnabled && len(p.Intervals) > 1 {
				for j := range p.Intervals {
					tier := &p.Intervals[j]
					model.LongContext = append(model.LongContext, desktopLongContextTier{
						FromTokens: tier.MinTokens,
						ToTokens:   tier.MaxTokens,
						Price: desktopPrice{
							Input:      desktopPerMillion(tier.InputPrice, money.charged, rate),
							Output:     desktopPerMillion(tier.OutputPrice, money.charged, rate),
							CacheRead:  desktopPerMillion(tier.CacheReadPrice, money.charged, rate),
							CacheWrite: desktopPerMillion(tier.CacheWritePrice, money.charged, rate),
						},
					})
				}
			}
		}
		if o := m.OfficialPricing; o != nil {
			model.Official = &desktopPrice{
				Input:      desktopPerMillion(o.InputPrice, money.official, 1),
				Output:     desktopPerMillion(o.OutputPrice, money.official, 1),
				CacheRead:  desktopPerMillion(o.CacheReadPrice, money.official, 1),
				CacheWrite: desktopPerMillion(o.CacheWritePrice, money.official, 1),
			}
		}
		if t := m.TimePricing; t != nil {
			for _, period := range t.Periods {
				model.Peak = append(model.Peak, desktopPeak{
					Start:        period.StartTime,
					End:          period.EndTime,
					Multiplier:   period.Multiplier,
					Timezone:     t.Timezone,
					WeekdaysOnly: t.WeekdaysOnly,
				})
			}
		}
		plan.Models = append(plan.Models, model)
	}
	return plan
}

// Plans handles GET /api/desktop/v2/plans
func (h *DesktopHandler) Plans(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	ctx := c.Request.Context()
	groups, err := h.plazaService.ListGroups(ctx)
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	// Only plans this user can put a key on: a plan that can be seen but not used is no choice.
	bindable, err := h.apiKeyService.GetAvailableGroups(ctx, subject.UserID)
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	usable := make(map[int64]struct{}, len(bindable))
	for i := range bindable {
		usable[bindable[i].ID] = struct{}{}
	}
	visible := make([]service.PlazaGroup, 0, len(groups))
	for i := range groups {
		if _, ok := usable[groups[i].ID]; ok {
			visible = append(visible, groups[i])
		}
	}
	// A price an admin set for this user wins over the growth-level price, as it does when billing.
	rates, _ := h.apiKeyService.GetUserGroupRates(ctx, subject.UserID)
	appLines := make([]*service.Group, 0, len(visible))
	for i := range visible {
		if visible[i].AppOnly {
			appLines = append(appLines, &service.Group{ID: visible[i].ID, AppOnly: true, RateMultiplier: visible[i].RateMultiplier})
		}
	}
	appRates := h.apiKeyService.AppLineRatesForUser(ctx, subject.UserID, appLines)

	money := h.money(c)
	plans := make([]desktopPlan, 0, len(visible))
	for i := range visible {
		g := &visible[i]
		rate := g.RateMultiplier
		if manual, ok := rates[g.ID]; ok {
			rate = manual
		} else if growth, ok := appRates[g.ID]; ok {
			rate = growth
		}
		plan := toDesktopPlan(g, rate, money)
		plan.Name, plan.Description = g.I18n.Localize(c.GetHeader("Accept-Language"), plan.Name, plan.Description)
		plans = append(plans, plan)
	}
	desktopOK(c, gin.H{"currency": money.Currency, "plans": plans})
}

func desktopKeyView(key *service.APIKey) gin.H {
	planID := ""
	if key.GroupID != nil {
		planID = strconv.FormatInt(*key.GroupID, 10)
	}
	return gin.H{
		"id":               strconv.FormatInt(key.ID, 10),
		"name":             key.Name,
		"planId":           planID,
		"status":           key.Status,
		"createdAtEpochMs": key.CreatedAt.UnixMilli(),
	}
}

// ListKeys handles GET /api/desktop/v2/keys
func (h *DesktopHandler) ListKeys(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	out := []gin.H{}
	for page := 1; ; page++ {
		keys, result, err := h.apiKeyService.List(c.Request.Context(), subject.UserID,
			pagination.PaginationParams{Page: page, PageSize: desktopMaxPageSize}, service.APIKeyListFilters{})
		if err != nil {
			desktopError(c, http.StatusInternalServerError, "internal_error")
			return
		}
		for i := range keys {
			out = append(out, desktopKeyView(&keys[i]))
		}
		if len(keys) == 0 || result == nil || int64(len(out)) >= result.Total {
			break
		}
	}
	desktopOK(c, gin.H{"keys": out})
}

type createDesktopKeyRequest struct {
	Name   string `json:"name"`
	PlanID string `json:"planId"`
}

// CreateKey handles POST /api/desktop/v2/keys
func (h *DesktopHandler) CreateKey(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	var req createDesktopKeyRequest
	_ = c.ShouldBindJSON(&req)
	req.Name = strings.TrimSpace(req.Name)
	planID, err := strconv.ParseInt(req.PlanID, 10, 64)
	if req.Name == "" || len(req.Name) > desktopKeyNameMaxLen || err != nil || planID <= 0 {
		desktopError(c, http.StatusBadRequest, "invalid_request")
		return
	}
	key, err := h.apiKeyService.Create(c.Request.Context(), subject.UserID, service.CreateAPIKeyRequest{Name: req.Name, GroupID: &planID})
	if errors.Is(err, service.ErrGroupNotAllowed) {
		desktopError(c, http.StatusForbidden, "plan_not_allowed")
		return
	}
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	view := desktopKeyView(key)
	view["key"] = key.Key
	desktopOK(c, view)
}

// ownedDesktopKey returns the caller's key named by the :id path parameter, or answers the request itself.
func (h *DesktopHandler) ownedDesktopKey(c *gin.Context) (*service.APIKey, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return nil, false
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		desktopError(c, http.StatusBadRequest, "invalid_request")
		return nil, false
	}
	key, err := h.apiKeyService.GetByID(c.Request.Context(), id)
	// A key that belongs to someone else is reported as missing, so ids cannot be probed.
	if errors.Is(err, service.ErrAPIKeyNotFound) || (err == nil && key.UserID != subject.UserID) {
		desktopError(c, http.StatusNotFound, "key_not_found")
		return nil, false
	}
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return nil, false
	}
	return key, true
}

// RevealKey handles POST /api/desktop/v2/keys/:id/reveal
func (h *DesktopHandler) RevealKey(c *gin.Context) {
	key, ok := h.ownedDesktopKey(c)
	if !ok {
		return
	}
	desktopOK(c, gin.H{"key": key.Key})
}

// DeleteKey handles DELETE /api/desktop/v2/keys/:id
func (h *DesktopHandler) DeleteKey(c *gin.Context) {
	key, ok := h.ownedDesktopKey(c)
	if !ok {
		return
	}
	if err := h.apiKeyService.Delete(c.Request.Context(), key.ID, key.UserID); err != nil && !errors.Is(err, service.ErrAPIKeyNotFound) {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	desktopOK(c, gin.H{})
}

// UsageSummary handles GET /api/desktop/v2/usage/summary
func (h *DesktopHandler) UsageSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	end := time.Now()
	stats, err := h.usageService.GetStatsByUser(c.Request.Context(), subject.UserID, end.AddDate(0, 0, -desktopUsageDays), end)
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	money := h.money(c)
	desktopOK(c, gin.H{
		"currency":   money.Currency,
		"windowDays": desktopUsageDays,
		"consumed":   money.charged(stats.TotalActualCost),
		"requests":   stats.TotalRequests,
		"tokens":     stats.TotalTokens,
	})
}

func desktopUsageRecord(log *service.UsageLog, money desktopMoney) gin.H {
	planID := ""
	if log.GroupID != nil {
		planID = strconv.FormatInt(*log.GroupID, 10)
	}
	keyName := ""
	if log.APIKey != nil {
		keyName = log.APIKey.Name
	}
	return gin.H{
		"id":           strconv.FormatInt(log.ID, 10),
		"atEpochMs":    log.CreatedAt.UnixMilli(),
		"model":        log.Model,
		"planId":       planID,
		"keyId":        strconv.FormatInt(log.APIKeyID, 10),
		"keyName":      keyName,
		"inputTokens":  log.InputTokens,
		"outputTokens": log.OutputTokens,
		"cacheTokens":  log.CacheReadTokens + log.CacheCreationTokens,
		"amount":       money.charged(log.ActualCost),
	}
}

func desktopPositiveQuery(c *gin.Context, name string, fallback, maximum int) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return min(value, maximum)
}

// UsageRecords handles GET /api/desktop/v2/usage/records
func (h *DesktopHandler) UsageRecords(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	page := desktopPositiveQuery(c, "page", 1, math.MaxInt32)
	pageSize := desktopPositiveQuery(c, "pageSize", desktopMaxPageSize, desktopMaxPageSize)
	logs, result, err := h.usageService.ListWithFilters(c.Request.Context(),
		pagination.PaginationParams{Page: page, PageSize: pageSize, SortBy: "created_at", SortOrder: "desc"},
		usagestats.UsageLogFilters{UserID: subject.UserID})
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	money := h.money(c)
	records := make([]gin.H, 0, len(logs))
	for i := range logs {
		records = append(records, desktopUsageRecord(&logs[i], money))
	}
	var total int64
	if result != nil {
		total = result.Total
	}
	desktopOK(c, gin.H{"currency": money.Currency, "records": records, "total": total, "page": page, "pageSize": pageSize})
}

func desktopNotice(a *service.UserAnnouncement) gin.H {
	published := a.Announcement.CreatedAt
	if a.Announcement.StartsAt != nil {
		published = *a.Announcement.StartsAt
	}
	var expires int64
	if a.Announcement.EndsAt != nil {
		expires = a.Announcement.EndsAt.UnixMilli()
	}
	return gin.H{
		"id":                 strconv.FormatInt(a.Announcement.ID, 10),
		"title":              a.Announcement.Title,
		"body":               a.Announcement.Content,
		"severity":           "info",
		"publishedAtEpochMs": published.UnixMilli(),
		"expiresAtEpochMs":   expires,
		"banner":             a.Announcement.NotifyMode == domain.AnnouncementNotifyModePopup,
	}
}

// Notices handles GET /api/desktop/v2/notices
func (h *DesktopHandler) Notices(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		desktopError(c, http.StatusUnauthorized, "session_required")
		return
	}
	items, err := h.announcementService.ListForUser(c.Request.Context(), subject.UserID, false)
	if err != nil {
		desktopError(c, http.StatusInternalServerError, "internal_error")
		return
	}
	notices := make([]gin.H, 0, len(items))
	for i := range items {
		notices = append(notices, desktopNotice(&items[i]))
	}
	desktopOK(c, gin.H{"notices": notices})
}
