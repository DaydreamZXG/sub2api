package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithOfficialCNY(t *testing.T) {
	usd := 1.5e-7
	cny := 1e-6
	catalog := &PlazaOfficialPricing{InputPrice: &usd}

	require.Same(t, catalog, withOfficialCNY(catalog, &ChannelModelPricing{}))
	require.Nil(t, withOfficialCNY(nil, nil))

	got := withOfficialCNY(catalog, &ChannelModelPricing{OfficialCNYInputPrice: &cny})
	require.Equal(t, &usd, got.InputPrice)
	require.Equal(t, &cny, got.CNYInputPrice)
	require.Nil(t, catalog.CNYInputPrice, "the shared catalog entry stays untouched")

	onlyCNY := withOfficialCNY(nil, &ChannelModelPricing{OfficialCNYOutputPrice: &cny})
	require.Nil(t, onlyCNY.InputPrice)
	require.Equal(t, &cny, onlyCNY.CNYOutputPrice)
}
