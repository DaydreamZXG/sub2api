package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

// SettingKeyOfficialPriceBook holds the site-wide official prices of models, shown on the price list
// next to every line that serves them. Display only: billing never reads it.
const SettingKeyOfficialPriceBook = "official_price_book"

const maxOfficialPriceBookEntries = 500

// OfficialPriceEntry is one model's official price, per million tokens. Models lists the names the model
// goes by on the lines; names match ignoring case and punctuation. A nil price is unknown.
type OfficialPriceEntry struct {
	Models       []string `json:"models"`
	CNYInput     *float64 `json:"cny_input"`
	CNYOutput    *float64 `json:"cny_output"`
	CNYCacheRead *float64 `json:"cny_cache_read"`
	USDInput     *float64 `json:"usd_input"`
	USDOutput    *float64 `json:"usd_output"`
	USDCacheRead *float64 `json:"usd_cache_read"`
}

// officialModelKey folds a model name so "GLM-5.3", "glm5.3" and "glm_5_3" are the same model.
func officialModelKey(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			_, _ = b.WriteRune(r)
		}
	}
	return b.String()
}

type officialPriceName struct {
	row  int
	name string
}

// NormalizeOfficialPriceBook trims names, drops empty ones and rejects entries the price list cannot use.
func NormalizeOfficialPriceBook(entries []OfficialPriceEntry) ([]OfficialPriceEntry, error) {
	if len(entries) > maxOfficialPriceBookEntries {
		return nil, fmt.Errorf("at most %d models", maxOfficialPriceBookEntries)
	}
	seen := map[string]officialPriceName{}
	out := make([]OfficialPriceEntry, 0, len(entries))
	for i, e := range entries {
		names := make([]string, 0, len(e.Models))
		for _, n := range e.Models {
			n = strings.TrimSpace(n)
			key := officialModelKey(n)
			if key == "" {
				continue
			}
			if first, dup := seen[key]; dup {
				if first.row == i {
					continue
				}
				return nil, fmt.Errorf("model %q is listed twice (also as %q)", n, first.name)
			}
			seen[key] = officialPriceName{row: i, name: n}
			names = append(names, n)
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("row %d has no model name", i+1)
		}
		for _, p := range []*float64{e.CNYInput, e.CNYOutput, e.CNYCacheRead, e.USDInput, e.USDOutput, e.USDCacheRead} {
			if p != nil && (*p < 0 || math.IsNaN(*p) || math.IsInf(*p, 0)) {
				return nil, fmt.Errorf("row %d has an invalid price", i+1)
			}
		}
		e.Models = names
		out = append(out, e)
	}
	return out, nil
}

// GetOfficialPriceBook returns the site-wide official prices; none saved yet is an empty book.
func (s *SettingService) GetOfficialPriceBook(ctx context.Context) ([]OfficialPriceEntry, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyOfficialPriceBook)
	if errors.Is(err, ErrSettingNotFound) || (err == nil && strings.TrimSpace(raw) == "") {
		return []OfficialPriceEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []OfficialPriceEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("parse official price book: %w", err)
	}
	return entries, nil
}

// UpdateOfficialPriceBook replaces the whole book.
func (s *SettingService) UpdateOfficialPriceBook(ctx context.Context, entries []OfficialPriceEntry) ([]OfficialPriceEntry, error) {
	normalized, err := NormalizeOfficialPriceBook(entries)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyOfficialPriceBook, string(raw)); err != nil {
		return nil, err
	}
	return normalized, nil
}

// officialPriceIndex maps each folded model name to its entry.
func officialPriceIndex(entries []OfficialPriceEntry) map[string]*OfficialPriceEntry {
	idx := make(map[string]*OfficialPriceEntry, len(entries))
	for i := range entries {
		for _, n := range entries[i].Models {
			if key := officialModelKey(n); key != "" {
				if _, taken := idx[key]; !taken {
					idx[key] = &entries[i]
				}
			}
		}
	}
	return idx
}

func perMillionToPerToken(v *float64) *float64 {
	if v == nil {
		return nil
	}
	t := *v / 1e6
	return &t
}

// withPriceBook lays the book's prices over the catalog's, without touching the shared catalog entry.
// Prices the book leaves empty keep the catalog's.
func withPriceBook(official *PlazaOfficialPricing, e *OfficialPriceEntry) *PlazaOfficialPricing {
	if e == nil {
		return official
	}
	var out PlazaOfficialPricing
	if official != nil {
		out = *official
	}
	set := func(dst **float64, v *float64) {
		if v != nil {
			*dst = perMillionToPerToken(v)
		}
	}
	set(&out.InputPrice, e.USDInput)
	set(&out.OutputPrice, e.USDOutput)
	set(&out.CacheReadPrice, e.USDCacheRead)
	set(&out.CNYInputPrice, e.CNYInput)
	set(&out.CNYOutputPrice, e.CNYOutput)
	set(&out.CNYCacheReadPrice, e.CNYCacheRead)
	if out.InputPrice == nil && out.OutputPrice == nil && out.CacheReadPrice == nil && out.CacheWritePrice == nil &&
		out.CNYInputPrice == nil && out.CNYOutputPrice == nil && out.CNYCacheReadPrice == nil && len(out.Intervals) == 0 {
		return official
	}
	return &out
}

// OfficialPriceFor returns the book's entry for a model, or nil when the book has none or cannot be read.
func (s *SettingService) OfficialPriceFor(ctx context.Context, model string) *OfficialPriceEntry {
	if s == nil {
		return nil
	}
	entries, err := s.GetOfficialPriceBook(ctx)
	if err != nil {
		return nil
	}
	return officialPriceIndex(entries)[officialModelKey(model)]
}
