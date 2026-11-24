// Package fxrate keeps a daily USD to CNY market rate. It is only used to restate official USD
// prices in yuan for comparison; nothing is billed with it.
package fxrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

// DefaultCNYPerUSD is used until the first market rate arrives.
const DefaultCNYPerUSD = 7.2

const (
	refreshEvery = 24 * time.Hour
	retryEvery   = 30 * time.Minute
	fetchTimeout = 10 * time.Second
	// A quote outside this range is a broken source, not the market.
	minSaneRate = 5.0
	maxSaneRate = 10.0
)

// Source is one public rate endpoint and how to read CNY per USD from its body.
type Source struct {
	URL   string
	Parse func(body []byte) (float64, error)
}

// DefaultSources are free, keyless endpoints, tried in order.
var DefaultSources = []Source{
	{URL: "https://api.frankfurter.app/latest?from=USD&to=CNY", Parse: parseRatesCNY},
	{URL: "https://open.er-api.com/v6/latest/USD", Parse: parseRatesCNY},
}

// Rates caches the last good rate and refreshes it in the background once it is a day old.
type Rates struct {
	mu        sync.Mutex
	rate      float64
	fetchedAt time.Time
	triedAt   time.Time
	fetching  bool

	client  *http.Client
	sources []Source
	now     func() time.Time
}

// New returns a cache reading the given sources.
func New(client *http.Client, sources []Source) *Rates {
	if client == nil {
		client = &http.Client{Timeout: fetchTimeout}
	}
	return &Rates{client: client, sources: sources, now: time.Now}
}

var std = New(nil, DefaultSources)

// CNYPerUSD returns the process-wide rate.
func CNYPerUSD() float64 { return std.CNYPerUSD() }

// CNYPerUSD returns the last good rate, or DefaultCNYPerUSD before one arrives, and starts a
// background refresh when the rate is stale. It never blocks on the network.
func (r *Rates) CNYPerUSD() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	stale := r.fetchedAt.IsZero() || now.Sub(r.fetchedAt) >= refreshEvery
	if stale && !r.fetching && (r.triedAt.IsZero() || now.Sub(r.triedAt) >= retryEvery) {
		r.fetching = true
		r.triedAt = now
		go r.refresh()
	}
	if r.rate > 0 {
		return r.rate
	}
	return DefaultCNYPerUSD
}

func (r *Rates) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*fetchTimeout)
	defer cancel()
	rate, err := r.Fetch(ctx)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fetching = false
	if err != nil {
		log.Printf("[fxrate] refresh failed, keeping %.4f: %v", r.rateOrDefault(), err)
		return
	}
	r.rate = rate
	r.fetchedAt = r.now()
}

func (r *Rates) rateOrDefault() float64 {
	if r.rate > 0 {
		return r.rate
	}
	return DefaultCNYPerUSD
}

// Fetch asks each source in turn and returns the first sane rate, rounded to 4 decimals.
func (r *Rates) Fetch(ctx context.Context) (float64, error) {
	var errs []error
	for _, s := range r.sources {
		rate, err := r.fetchOne(ctx, s)
		if err == nil {
			return math.Round(rate*10000) / 10000, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", s.URL, err))
	}
	return 0, errors.Join(errs...)
}

func (r *Rates) fetchOne(ctx context.Context, s Source) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	var body json.RawMessage
	if err := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, 1<<20)).Decode(&body); err != nil {
		return 0, err
	}
	rate, err := s.Parse(body)
	if err != nil {
		return 0, err
	}
	if rate < minSaneRate || rate > maxSaneRate {
		return 0, fmt.Errorf("rate %v out of range", rate)
	}
	return rate, nil
}

// parseRatesCNY reads {"rates":{"CNY":7.1}}, the shape both default sources use.
func parseRatesCNY(body []byte) (float64, error) {
	var v struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return 0, err
	}
	rate, ok := v.Rates["CNY"]
	if !ok {
		return 0, errors.New("no CNY rate")
	}
	return rate, nil
}
