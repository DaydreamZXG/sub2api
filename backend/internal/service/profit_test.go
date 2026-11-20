//go:build unit

package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCostPackCostBetween(t *testing.T) {
	oneOff := CostPack{Payments: []CostPackPayment{
		{Amount: 36, PaidAt: "2026-03-10"},
		{Amount: 20, PaidAt: "2026-04-02"},
	}}
	require.InDelta(t, 36, oneOff.costBetween("2026-03-01", "2026-03-31", "2026-06-01"), 1e-9)
	require.InDelta(t, 56, oneOff.costBetween("", "2026-06-01", "2026-06-01"), 1e-9)
	require.InDelta(t, 0, oneOff.costBetween("2026-05-01", "2026-05-31", "2026-06-01"), 1e-9)
	// A payment dated after today is not spent yet.
	require.InDelta(t, 36, oneOff.costBetween("", "2026-12-31", "2026-03-20"), 1e-9)

	monthly := CostPack{CycleDays: 30, CycleAmount: 150, StartedAt: "2026-01-01"}
	// Cycles start on 01-01, 01-31, 03-02, 04-01.
	require.InDelta(t, 300, monthly.costBetween("2026-01-01", "2026-01-31", "2026-06-01"), 1e-9)
	require.InDelta(t, 0, monthly.costBetween("2026-02-01", "2026-02-28", "2026-06-01"), 1e-9)
	require.InDelta(t, 600, monthly.costBetween("", "2026-04-01", "2026-06-01"), 1e-9)
	// Only cycles that have started by today count.
	require.InDelta(t, 150, monthly.costBetween("", "2026-12-31", "2026-01-15"), 1e-9)

	ended := CostPack{CycleDays: 30, CycleAmount: 150, StartedAt: "2026-01-01", EndedAt: strPtr("2026-02-15")}
	require.InDelta(t, 300, ended.costBetween("", "2026-12-31", "2026-12-31"), 1e-9)
}

func TestCostPackNextCycle(t *testing.T) {
	monthly := CostPack{CycleDays: 30, CycleAmount: 150, StartedAt: "2026-01-01"}
	next := monthly.nextCycle("2026-01-31")
	require.NotNil(t, next)
	require.Equal(t, "2026-03-02", *next)

	monthly.EndedAt = strPtr("2026-02-15")
	require.Nil(t, monthly.nextCycle("2026-01-31"))
	require.Nil(t, (&CostPack{StartedAt: "2026-01-01"}).nextCycle("2026-01-31"))
}

func TestNormalizeCostPackInput(t *testing.T) {
	ok, err := normalizeCostPackInput(CostPackInput{
		Name: "  Claude Max  ", StartedAt: "2026-03-01", EndedAt: strPtr(" "), CycleDays: 30, CycleAmount: 1400,
		Amount: 5, AccountIDs: []int64{3, 3, 7},
	}, true)
	require.NoError(t, err)
	require.Equal(t, "Claude Max", ok.Name)
	require.Nil(t, ok.EndedAt)
	require.Equal(t, []int64{3, 7}, ok.AccountIDs)
	// A subscription has no purchase price of its own.
	require.Zero(t, ok.Amount)

	bad := []CostPackInput{
		{Name: "", StartedAt: "2026-03-01", Amount: 5},
		{Name: "x", StartedAt: "03/01", Amount: 5},
		{Name: "x", StartedAt: "2026-03-01", Amount: 0},
		{Name: "x", StartedAt: "2026-03-01", CycleDays: 30},
		{Name: "x", StartedAt: "2026-03-01", CycleDays: -1, Amount: 5},
		{Name: "x", StartedAt: "2026-03-01", EndedAt: strPtr("2026-02-01"), Amount: 5},
		{Name: "x", StartedAt: "2026-03-01", Amount: 5, AccountIDs: []int64{0}},
	}
	for i, in := range bad {
		_, err := normalizeCostPackInput(in, true)
		require.ErrorIs(t, err, ErrCostPackInvalid, "case %d", i)
	}

	// Changing a one-off pack does not ask for its price again.
	_, err = normalizeCostPackInput(CostPackInput{Name: "x", StartedAt: "2026-03-01"}, false)
	require.NoError(t, err)
}

func TestBuildProfitReport(t *testing.T) {
	packs := []CostPack{
		{ID: 1, Name: "5 for 100", StartedAt: "2026-03-01", Payments: []CostPackPayment{{Amount: 36, PaidAt: "2026-03-01"}}},
		{ID: 2, Name: "Max", StartedAt: "2026-03-01", CycleDays: 30, CycleAmount: 100},
		{ID: 3, Name: "Idle", StartedAt: "2026-04-05", Payments: []CostPackPayment{{Amount: 10, PaidAt: "2026-04-05"}}},
	}
	rows := []ProfitUsageRow{
		// Account 10 is in pack 1 and serves two groups: 3/4 of its standard use on group 1, 1/4 on group 2.
		{Month: "2026-03", AccountID: 10, GroupID: 1, PackID: 1, Requests: 30, StandardCost: 300, ActualCost: 120, AccountCost: 300, PaidCost: 120},
		{Month: "2026-03", AccountID: 10, GroupID: 2, PackID: 1, Requests: 10, StandardCost: 100, ActualCost: 40, AccountCost: 100, PaidCost: 40},
		// Account 20 is in pack 2; a monthly-plan user used it, which is not money received.
		{Month: "2026-04", AccountID: 20, GroupID: 1, PackID: 2, BillingType: BillingTypeSubscription, Requests: 5, StandardCost: 50, ActualCost: 50, AccountCost: 50, PaidCost: 50},
		{Month: "2026-04", AccountID: 20, GroupID: 1, PackID: 2, Requests: 8, StandardCost: 80, ActualCost: 64, AccountCost: 80, PaidCost: 64},
		// Account 30 is paid by use at a rate of 0.3.
		{Month: "2026-04", AccountID: 30, GroupID: 2, Requests: 4, StandardCost: 100, ActualCost: 80, AccountCost: 30, PaidCost: 80},
	}

	// Two balance units per yuan: every usage amount is halved.
	report := buildProfitReport(rows, packs, "2026-03-01", "2026-04-30", "2026-04-30", 2)

	// Income: (120 + 40 + 64 + 80) / 2. Plan usage: 50 / 2. Rate cost: only account 30, 30 / 2.
	require.InDelta(t, 152, report.Totals.Income, 1e-9)
	require.InDelta(t, 25, report.Totals.PlanUsage, 1e-9)
	require.InDelta(t, 15, report.Totals.RateCost, 1e-9)
	// Packs: 36 + three cycles of 100 (03-01, 03-31, 04-30) + 10.
	require.InDelta(t, 346, report.Totals.PackCost, 1e-9)
	require.InDelta(t, 152-15-346, report.Totals.Profit, 1e-9)
	require.EqualValues(t, 57, report.Totals.Requests)

	require.Len(t, report.Months, 2)
	require.Equal(t, "2026-03", report.Months[0].Month)
	require.InDelta(t, 80, report.Months[0].Income, 1e-9)
	require.InDelta(t, 236, report.Months[0].PackCost, 1e-9)
	require.InDelta(t, 110, report.Months[1].PackCost, 1e-9)
	require.InDelta(t, 15, report.Months[1].RateCost, 1e-9)

	groups := map[int64]ProfitGroupLine{}
	for _, g := range report.Groups {
		groups[g.GroupID] = g
	}
	// Pack 1 splits 36 as 27 / 9; pack 2 falls on group 1 alone; pack 3 was never used.
	require.InDelta(t, 27+300, groups[1].PackCost, 1e-9)
	require.InDelta(t, 9, groups[2].PackCost, 1e-9)
	require.InDelta(t, 15, groups[2].RateCost, 1e-9)
	require.InDelta(t, 10, groups[profitUnallocatedID].PackCost, 1e-9)
	require.Equal(t, int64(profitUnallocatedID), report.Groups[len(report.Groups)-1].GroupID)
	groupCost := 0.0
	for _, g := range report.Groups {
		groupCost += g.PackCost
	}
	require.InDelta(t, report.Totals.PackCost, groupCost, 1e-9)

	packLines := map[int64]ProfitPackLine{}
	for _, p := range report.Packs {
		packLines[p.PackID] = p
	}
	require.InDelta(t, 80-36, packLines[1].Profit, 1e-9)
	require.InDelta(t, 32-300, packLines[2].Profit, 1e-9)
	require.InDelta(t, 25, packLines[2].PlanUsage, 1e-9)
	require.InDelta(t, -10, packLines[3].Profit, 1e-9)
	// The worst pack comes first.
	require.Equal(t, int64(2), report.Packs[0].PackID)

	accounts := map[int64]ProfitAccountLine{}
	for _, a := range report.Accounts {
		accounts[a.AccountID] = a
	}
	require.Equal(t, int64(1), accounts[10].PackID)
	require.Equal(t, "5 for 100", accounts[10].PackName)
	// An account under a pack carries no cost by use.
	require.Zero(t, accounts[10].RateCost)
	require.Zero(t, accounts[30].PackID)
	require.InDelta(t, 40-15, accounts[30].Profit, 1e-9)
}

func TestBuildProfitReportGiftedBalanceIsNotIncome(t *testing.T) {
	rows := []ProfitUsageRow{
		// 100 charged, 60 of it real money; the account costs 30 by use.
		{Month: "2026-04", AccountID: 30, GroupID: 2, Requests: 4, StandardCost: 100, ActualCost: 100, AccountCost: 30, PaidCost: 60},
		// A monthly plan's use is neither income nor gifted.
		{Month: "2026-04", AccountID: 30, GroupID: 2, BillingType: BillingTypeSubscription, Requests: 1, StandardCost: 10, ActualCost: 10, AccountCost: 3, PaidCost: 6},
	}
	report := buildProfitReport(rows, nil, "2026-04-01", "2026-04-30", "2026-04-30", 2)

	require.InDelta(t, 30, report.Totals.Income, 1e-9)
	require.InDelta(t, 20, report.Totals.Gifted, 1e-9)
	require.InDelta(t, 5, report.Totals.PlanUsage, 1e-9)
	require.InDelta(t, 16.5, report.Totals.RateCost, 1e-9)
	require.InDelta(t, 13.5, report.Totals.Profit, 1e-9)
	require.Len(t, report.Groups, 1)
	require.InDelta(t, 20, report.Groups[0].Gifted, 1e-9)
	require.Len(t, report.Accounts, 1)
	require.InDelta(t, 30, report.Accounts[0].Income, 1e-9)
}

func TestBuildProfitReportEmpty(t *testing.T) {
	report := buildProfitReport(nil, nil, "2026-03-01", "2026-03-31", "2026-03-31", 0)
	require.Equal(t, 1.0, report.UnitsPerYuan)
	require.Zero(t, report.Totals.Profit)

	// A period with nothing in it is sent as empty lists; the page reads their length.
	body, err := json.Marshal(report)
	require.NoError(t, err)
	for _, list := range []string{"months", "groups", "packs", "accounts"} {
		require.Contains(t, string(body), `"`+list+`":[]`)
	}
}
