package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func desktopFloat(v float64) *float64 { return &v }

func TestDesktopMoneyUsesTheCurrencyUsersPayIn(t *testing.T) {
	usd := newDesktopMoney(&service.PublicSettings{PaymentEnabled: false, BalanceRechargeMultiplier: 10})
	require.Equal(t, "USD", usd.Currency)
	require.Equal(t, 12.5, usd.charged(12.5))
	require.Equal(t, 3.0, usd.official(3))

	// 1 yuan buys 2 balance units; official prices are restated at 7 yuan per dollar.
	cny := newDesktopMoney(&service.PublicSettings{PaymentEnabled: true, BalanceRechargeMultiplier: 2, SubscriptionUSDToCNYRate: 7})
	require.Equal(t, "CNY", cny.Currency)
	require.Equal(t, 6.25, cny.charged(12.5))
	require.Equal(t, 21.0, cny.official(3))

	fallbackRate := newDesktopMoney(&service.PublicSettings{PaymentEnabled: true, BalanceRechargeMultiplier: 1})
	require.Equal(t, 7.2, fallbackRate.official(1))

	require.Equal(t, "USD", newDesktopMoney(nil).Currency)
}

func TestToDesktopPlanPricesPerMillionTokensWithSurchargesKeptApart(t *testing.T) {
	maxTokens := 200_000
	group := &service.PlazaGroup{
		ID:                        7,
		Name:                      "Garden",
		Platform:                  domain.PlatformAnthropic,
		RateMultiplier:            1,
		AppOnly:                   true,
		AppFallback:               true,
		PeakRateEnabled:           true,
		PeakStart:                 "09:00",
		PeakEnd:                   "18:00",
		PeakRateMultiplier:        1.5,
		LongContextPricingEnabled: true,
		Models: []service.PlazaModel{{
			Name: "model-a",
			Pricing: &service.ChannelModelPricing{
				InputPrice:  desktopFloat(0.000003),
				OutputPrice: desktopFloat(0.000015),
				Intervals: []service.PricingInterval{
					{MinTokens: 0, MaxTokens: &maxTokens, InputPrice: desktopFloat(0.000003)},
					{MinTokens: 200_000, InputPrice: desktopFloat(0.000006)},
				},
			},
			OfficialPricing: &service.PlazaOfficialPricing{InputPrice: desktopFloat(0.000003)},
			TimePricing: &service.TimePricingSchedule{
				Timezone: "Asia/Shanghai",
				Periods:  []service.TimePricingPeriod{{StartTime: "00:00", EndTime: "08:00", Multiplier: 0.5}},
			},
		}},
	}
	money := desktopMoney{Currency: "CNY", UnitsPerYuan: 1, CNYPerUSD: 7}

	plan := toDesktopPlan(group, 0.5, money)

	require.Equal(t, "7", plan.ID)
	require.Equal(t, []string{"anthropic"}, plan.Formats)
	require.True(t, plan.AppOnly)
	require.True(t, plan.AppFallback)
	require.False(t, plan.Subscription)
	require.Equal(t, 0.5, plan.Multiplier)
	require.Equal(t, []desktopPeak{{Start: "09:00", End: "18:00", Multiplier: 1.5}}, plan.Peak)

	require.Len(t, plan.Models, 1)
	model := plan.Models[0]
	require.Equal(t, 1.5, *model.Price.Input)
	require.Equal(t, 7.5, *model.Price.Output)
	require.Nil(t, model.Price.CacheRead)
	require.Equal(t, 21.0, *model.Official.Input)
	require.Len(t, model.LongContext, 2)
	require.Equal(t, 200_000, *model.LongContext[0].ToTokens)
	require.Equal(t, 200_000, model.LongContext[1].FromTokens)
	require.Nil(t, model.LongContext[1].ToTokens)
	require.Equal(t, 3.0, *model.LongContext[1].Price.Input)
	require.Equal(t, []desktopPeak{{Start: "00:00", End: "08:00", Multiplier: 0.5, Timezone: "Asia/Shanghai"}}, model.Peak)
}

func TestToDesktopPlanLeavesOutWhatDoesNotApply(t *testing.T) {
	group := &service.PlazaGroup{
		ID:                 3,
		Platform:           "something-else",
		SubscriptionType:   service.SubscriptionTypeSubscription,
		AppFallback:        true,
		PeakRateEnabled:    true,
		PeakRateMultiplier: 1,
		Models: []service.PlazaModel{{
			Name: "per-call",
			Pricing: &service.ChannelModelPricing{
				PerRequestPrice: desktopFloat(0.02),
				Intervals:       []service.PricingInterval{{MinTokens: 0}, {MinTokens: 100}},
			},
		}},
	}

	plan := toDesktopPlan(group, 2, desktopMoney{Currency: "USD"})

	require.True(t, plan.Subscription)
	require.False(t, plan.AppFallback, "a fallback only means something on an app-only plan")
	require.Empty(t, plan.Formats)
	require.NotNil(t, plan.Formats, "unknown formats are an empty list, not null")
	require.Nil(t, plan.Peak, "a peak multiplier of 1 is not a surcharge")
	require.Equal(t, 0.04, *plan.Models[0].Price.PerRequest)
	require.Nil(t, plan.Models[0].LongContext, "tiers are not billed unless the plan turns them on")
	require.Nil(t, plan.Models[0].Official)
}

func TestDesktopNoticeMapsAnnouncementToTheProtocolShape(t *testing.T) {
	created := time.UnixMilli(1_700_000_000_000)
	starts := time.UnixMilli(1_700_000_100_000)
	ends := time.UnixMilli(1_700_000_200_000)

	popup := desktopNotice(&service.UserAnnouncement{Announcement: service.Announcement{
		ID: 9, Title: "T", Content: "B", NotifyMode: domain.AnnouncementNotifyModePopup,
		StartsAt: &starts, EndsAt: &ends, CreatedAt: created,
	}})
	require.Equal(t, gin.H{
		"id": "9", "title": "T", "body": "B", "severity": "info",
		"publishedAtEpochMs": int64(1_700_000_100_000), "expiresAtEpochMs": int64(1_700_000_200_000), "banner": true,
	}, popup)

	silent := desktopNotice(&service.UserAnnouncement{Announcement: service.Announcement{
		ID: 10, Title: "T", NotifyMode: domain.AnnouncementNotifyModeSilent, CreatedAt: created,
	}})
	require.Equal(t, int64(1_700_000_000_000), silent["publishedAtEpochMs"])
	require.Equal(t, int64(0), silent["expiresAtEpochMs"])
	require.Equal(t, false, silent["banner"])
}

func TestDesktopUsageRecordConvertsTheAmountCharged(t *testing.T) {
	groupID := int64(4)
	record := desktopUsageRecord(&service.UsageLog{
		ID: 11, APIKeyID: 5, Model: "model-a", GroupID: &groupID,
		InputTokens: 100, OutputTokens: 20, CacheReadTokens: 7, CacheCreationTokens: 3,
		TotalCost: 9, ActualCost: 3, CreatedAt: time.UnixMilli(1_700_000_000_000),
		APIKey: &service.APIKey{Name: "Codex"},
	}, desktopMoney{Currency: "CNY", UnitsPerYuan: 2, CNYPerUSD: 7})

	require.Equal(t, gin.H{
		"id": "11", "atEpochMs": int64(1_700_000_000_000), "model": "model-a", "planId": "4",
		"keyId": "5", "keyName": "Codex", "inputTokens": 100, "outputTokens": 20, "cacheTokens": 10, "amount": 1.5,
	}, record)
}

func TestDesktopBootstrapDescribesTheStation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/desktop/v2/bootstrap", (&DesktopHandler{}).Bootstrap)

	req := httptest.NewRequest(http.MethodGet, "/api/desktop/v2/bootstrap", nil)
	req.Host = "relay.example.com"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	var body struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "desktop-connect", body.Data["service"])
	require.Equal(t, "desktop-connect-v1", body.Data["contract_id"])
	require.Equal(t, float64(2), body.Data["schema_version"])
	require.Equal(t, "https://relay.example.com/topup", body.Data["wallet_url"])
	require.Equal(t, "X-Desktop-App-Token", body.Data["app_token_header"])
	for key, value := range body.Data {
		if strings.HasSuffix(key, "_path") {
			path, ok := value.(string)
			require.Truef(t, ok && strings.HasPrefix(path, "/api/desktop/v2/"), "%s = %v", key, value)
		}
	}
	require.NotContains(t, strings.ToLower(rec.Body.String()), "yeschoy", "the protocol names no particular station")
}

func TestDesktopRefreshSessionRejectsIncompleteRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", (&DesktopHandler{}).RefreshSession)

	for _, payload := range []string{``, `{}`, `{"refresh_token":"rt"}`, `{"session_id":"s"}`} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(payload)))
		require.Equal(t, http.StatusBadRequest, rec.Code, payload)
		require.JSONEq(t, `{"success":false,"code":"invalid_request"}`, rec.Body.String())
	}
}
