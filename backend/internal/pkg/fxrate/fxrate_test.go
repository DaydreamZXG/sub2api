package fxrate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func serve(t *testing.T, status int, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestFetchFallsBackToNextSourceAndRounds(t *testing.T) {
	down := serve(t, http.StatusBadGateway, "")
	broken := serve(t, http.StatusOK, `{"rates":{"CNY":71.2}}`)
	good := serve(t, http.StatusOK, `{"result":"success","rates":{"CNY":7.123456,"EUR":0.9}}`)
	r := New(nil, []Source{{URL: down, Parse: parseRatesCNY}, {URL: broken, Parse: parseRatesCNY}, {URL: good, Parse: parseRatesCNY}})

	rate, err := r.Fetch(context.Background())
	require.NoError(t, err)
	require.Equal(t, 7.1235, rate)
}

func TestFetchFailsWhenEverySourceFails(t *testing.T) {
	r := New(nil, []Source{{URL: serve(t, http.StatusOK, `{"rates":{}}`), Parse: parseRatesCNY}})
	_, err := r.Fetch(context.Background())
	require.Error(t, err)
}

func TestCNYPerUSDServesDefaultThenMarketRate(t *testing.T) {
	r := New(nil, []Source{{URL: serve(t, http.StatusOK, `{"rates":{"CNY":7.05}}`), Parse: parseRatesCNY}})

	require.Equal(t, DefaultCNYPerUSD, r.CNYPerUSD())
	require.Eventually(t, func() bool { return r.CNYPerUSD() == 7.05 }, 5*time.Second, 10*time.Millisecond)
}

func TestCNYPerUSDKeepsLastRateWhenRefreshFails(t *testing.T) {
	r := New(nil, []Source{{URL: serve(t, http.StatusInternalServerError, ""), Parse: parseRatesCNY}})
	now := time.Now()
	r.now = func() time.Time { return now }
	r.rate = 7.01
	r.fetchedAt = now.Add(-25 * time.Hour)

	require.Equal(t, 7.01, r.CNYPerUSD())
	require.Eventually(t, func() bool {
		r.mu.Lock()
		defer r.mu.Unlock()
		return !r.fetching
	}, 5*time.Second, 10*time.Millisecond)
	require.Equal(t, 7.01, r.CNYPerUSD())
}

func TestRefreshNotifiesOnlyWhenTheRateChanges(t *testing.T) {
	r := New(nil, []Source{{URL: serve(t, http.StatusOK, `{"rates":{"CNY":6.66}}`), Parse: parseRatesCNY}})
	calls := make(chan struct{}, 4)
	r.OnChange(func() { calls <- struct{}{} })

	r.refresh()
	require.Len(t, calls, 1)
	r.refresh()
	require.Len(t, calls, 1)
}
