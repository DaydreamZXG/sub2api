//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func bookFloat(v float64) *float64 { return &v }

func TestOfficialModelKeyIgnoresCaseAndPunctuation(t *testing.T) {
	require.Equal(t, "glm53", officialModelKey("GLM-5.3"))
	require.Equal(t, officialModelKey("glm5.3"), officialModelKey("glm_5_3"))
	require.NotEqual(t, officialModelKey("glm-5.3"), officialModelKey("glm-5.3-flash"))
}

func TestNormalizeOfficialPriceBookRejectsWhatThePriceListCannotUse(t *testing.T) {
	_, err := NormalizeOfficialPriceBook([]OfficialPriceEntry{{Models: []string{" ", "--"}}})
	require.Error(t, err, "a row needs a model name")

	_, err = NormalizeOfficialPriceBook([]OfficialPriceEntry{{Models: []string{"GLM-5.3"}}, {Models: []string{"glm5.3"}}})
	require.Error(t, err, "the same model in two rows")

	_, err = NormalizeOfficialPriceBook([]OfficialPriceEntry{{Models: []string{"m"}, CNYInput: bookFloat(-1)}})
	require.Error(t, err)

	out, err := NormalizeOfficialPriceBook([]OfficialPriceEntry{{Models: []string{" deepseek-flash ", "", "DeepSeek-Flash"}, CNYInput: bookFloat(1)}})
	require.NoError(t, err, "the same name twice in one row is harmless")
	require.Equal(t, []string{"deepseek-flash"}, out[0].Models)
}

func TestWithPriceBookLaysPricesOverTheCatalog(t *testing.T) {
	catalog := &PlazaOfficialPricing{InputPrice: bookFloat(1e-6), CacheWritePrice: bookFloat(2e-6)}
	entry := &OfficialPriceEntry{USDOutput: bookFloat(4), CNYInput: bookFloat(7)}

	got := withPriceBook(catalog, entry)

	require.Equal(t, 1e-6, *got.InputPrice, "prices the book leaves empty keep the catalog's")
	require.Equal(t, 2e-6, *got.CacheWritePrice)
	require.InDelta(t, 4e-6, *got.OutputPrice, 1e-18)
	require.InDelta(t, 7e-6, *got.CNYInputPrice, 1e-18)
	require.Nil(t, catalog.OutputPrice, "the shared catalog entry is not touched")
}

func TestWithPriceBookGivesModelsTheCatalogDoesNotKnow(t *testing.T) {
	require.Nil(t, withPriceBook(nil, nil))
	require.Nil(t, withPriceBook(nil, &OfficialPriceEntry{Models: []string{"m"}}), "an entry with no prices adds nothing")
	got := withPriceBook(nil, &OfficialPriceEntry{CNYInput: bookFloat(8)})
	require.InDelta(t, 8e-6, *got.CNYInputPrice, 1e-18)
}

func TestLinePricesWinOverTheBook(t *testing.T) {
	idx := officialPriceIndex([]OfficialPriceEntry{{Models: []string{"GLM-5.3"}, CNYInput: bookFloat(8), CNYOutput: bookFloat(28)}})
	fromBook := withPriceBook(nil, idx[officialModelKey("glm-5.3")])
	got := withOfficialCNY(fromBook, &ChannelModelPricing{OfficialCNYInputPrice: bookFloat(5e-6)})
	require.Equal(t, 5e-6, *got.CNYInputPrice)
	require.InDelta(t, 28e-6, *got.CNYOutputPrice, 1e-18, "a price the line leaves empty keeps the book's")
}
