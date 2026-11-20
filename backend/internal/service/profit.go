package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

// SettingKeyProfitBackfilled is set once the usage that predates the profit summary has been folded into it.
const SettingKeyProfitBackfilled = "profit_backfilled"

const (
	profitDateLayout     = "2006-01-02"
	profitMaxAmount      = 1e9
	profitMaxCycleDays   = 3660
	profitMaxReportDays  = 1100
	profitUnallocatedID  = -1
	costPackNameMaxRunes = 100
)

var (
	ErrCostPackInvalid      = infraerrors.BadRequest("COST_PACK_INVALID", "invalid cost pack")
	ErrCostPackNotFound     = infraerrors.NotFound("COST_PACK_NOT_FOUND", "cost pack not found")
	ErrCostPackAccountTaken = infraerrors.Conflict("COST_PACK_ACCOUNT_TAKEN", "an account already belongs to another cost pack")
	ErrProfitRangeInvalid   = infraerrors.BadRequest("PROFIT_RANGE_INVALID", "invalid date range")
)

// CostPack is money actually paid for upstream capacity, tied to the accounts it bought.
// Dates are calendar days ("2006-01-02") in the site's timezone.
type CostPack struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Note string `json:"note"`
	// CycleDays 0 is a one-off pack; above 0, CycleAmount is paid again every CycleDays (a subscription).
	CycleDays   int     `json:"cycle_days"`
	CycleAmount float64 `json:"cycle_amount"`
	StartedAt   string  `json:"started_at"`
	// EndedAt is the last day the pack counts: no cycle starts and no income is credited to it after that day.
	EndedAt  *string           `json:"ended_at"`
	Payments []CostPackPayment `json:"payments"`
	Accounts []CostPackAccount `json:"accounts"`

	// The figures below are worked out for each response, in yuan.
	Paid      float64 `json:"paid"`
	Income    float64 `json:"income"`
	PlanUsage float64 `json:"plan_usage"`
	Profit    float64 `json:"profit"`
	Requests  int64   `json:"requests"`
	Ended     bool    `json:"ended"`
	// NextCycleAt is the day the next cycle of a running subscription falls due.
	NextCycleAt *string `json:"next_cycle_at"`
}

// CostPackPayment is one-off money put into a pack: the purchase itself or a later top-up.
type CostPackPayment struct {
	ID     int64   `json:"id"`
	Amount float64 `json:"amount"`
	PaidAt string  `json:"paid_at"`
	Note   string  `json:"note"`
}

// CostPackAccount is an upstream account a pack pays for.
type CostPackAccount struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	Deleted  bool   `json:"deleted"`
}

// CostPackInput is what an admin fills in to create or change a pack.
type CostPackInput struct {
	Name        string  `json:"name"`
	Note        string  `json:"note"`
	CycleDays   int     `json:"cycle_days"`
	CycleAmount float64 `json:"cycle_amount"`
	StartedAt   string  `json:"started_at"`
	EndedAt     *string `json:"ended_at"`
	AccountIDs  []int64 `json:"account_ids"`
	// Amount is the purchase price of a one-off pack; it is only read when the pack is created.
	Amount float64 `json:"amount"`
}

// ProfitUsageRow is usage summed for one month, account, group and billing type.
// PackID is the pack the account belonged to on those days, 0 when it was outside any pack.
type ProfitUsageRow struct {
	Month        string
	AccountID    int64
	GroupID      int64
	BillingType  int8
	PackID       int64
	Requests     int64
	StandardCost float64
	ActualCost   float64
	AccountCost  float64
}

// ProfitPackUsage is what a pack's accounts earned while they were in it.
type ProfitPackUsage struct {
	Requests  int64
	Income    float64
	PlanUsage float64
}

// ProfitAccountInfo names an upstream account for the report.
type ProfitAccountInfo struct {
	Name           string
	Platform       string
	RateMultiplier float64
	Deleted        bool
}

// ProfitRepository stores cost packs and reads the hourly usage summary.
type ProfitRepository interface {
	ListPacks(ctx context.Context) ([]CostPack, error)
	GetPack(ctx context.Context, id int64) (*CostPack, error)
	CreatePack(ctx context.Context, in CostPackInput) (int64, error)
	UpdatePack(ctx context.Context, id int64, in CostPackInput) error
	DeletePack(ctx context.Context, id int64) error
	AddPayment(ctx context.Context, packID int64, amount float64, paidAt, note string) error
	DeletePayment(ctx context.Context, packID, paymentID int64) error
	// Usage sums the summary between two days (both included), in the given timezone.
	Usage(ctx context.Context, from, to, tz string) ([]ProfitUsageRow, error)
	PackUsage(ctx context.Context, tz string) (map[int64]ProfitPackUsage, error)
	AccountInfos(ctx context.Context, ids []int64) (map[int64]ProfitAccountInfo, error)
	GroupNames(ctx context.Context, ids []int64) (map[int64]string, error)
	// RebuildSummary folds every usage log still kept into the summary.
	RebuildSummary(ctx context.Context) error
}

// ProfitFigures are the numbers every line of the report carries, in yuan.
type ProfitFigures struct {
	Requests int64 `json:"requests"`
	// Income is what users were charged from their balance.
	Income float64 `json:"income"`
	// PlanUsage is what monthly-plan users used, at plan value. It is not money received.
	PlanUsage float64 `json:"plan_usage"`
	// RateCost is the cost of accounts paid by use: standard price times the account's rate.
	RateCost float64 `json:"rate_cost"`
	// PackCost is the cost packs paid in the period, or the share of it that falls on this line.
	PackCost float64 `json:"pack_cost"`
	Profit   float64 `json:"profit"`
}

func (f *ProfitFigures) settle() {
	f.Profit = f.Income - f.RateCost - f.PackCost
}

type ProfitMonthLine struct {
	Month string `json:"month"`
	ProfitFigures
}

// ProfitGroupLine is one line (group) of the site. ID -1 holds pack cost that no usage could carry.
type ProfitGroupLine struct {
	GroupID int64  `json:"group_id"`
	Name    string `json:"name"`
	ProfitFigures
}

type ProfitPackLine struct {
	PackID int64  `json:"pack_id"`
	Name   string `json:"name"`
	ProfitFigures
}

type ProfitAccountLine struct {
	AccountID int64  `json:"account_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	Deleted   bool   `json:"deleted"`
	// PackID and PackName are set when a pack pays for this account; otherwise RateMultiplier is its cost rate.
	PackID         int64   `json:"pack_id"`
	PackName       string  `json:"pack_name"`
	RateMultiplier float64 `json:"rate_multiplier"`
	ProfitFigures
}

// ProfitReport is income, cost and profit between two days, in yuan.
type ProfitReport struct {
	From         string              `json:"from"`
	To           string              `json:"to"`
	UnitsPerYuan float64             `json:"units_per_yuan"`
	Totals       ProfitFigures       `json:"totals"`
	Months       []ProfitMonthLine   `json:"months"`
	Groups       []ProfitGroupLine   `json:"groups"`
	Packs        []ProfitPackLine    `json:"packs"`
	Accounts     []ProfitAccountLine `json:"accounts"`
}

// ProfitService works out what the site earned against what its upstream capacity cost.
type ProfitService struct {
	repo        ProfitRepository
	settingRepo SettingRepository
	payment     *PaymentConfigService
	backfill    sync.Mutex
	backfilled  bool
	now         func() time.Time
}

func NewProfitService(repo ProfitRepository, settingRepo SettingRepository, payment *PaymentConfigService) *ProfitService {
	return &ProfitService{repo: repo, settingRepo: settingRepo, payment: payment, now: timezone.Now}
}

func (s *ProfitService) today() string {
	return s.now().In(timezone.Location()).Format(profitDateLayout)
}

// unitsPerYuan is how much balance one yuan buys; amounts in balance units are divided by it to get yuan.
func (s *ProfitService) unitsPerYuan(ctx context.Context) float64 {
	if s.payment == nil || !s.payment.IsPaymentEnabled(ctx) {
		return defaultBalanceRechargeMultiplier
	}
	cfg, err := s.payment.GetPaymentConfig(ctx)
	if err != nil || cfg == nil {
		return defaultBalanceRechargeMultiplier
	}
	return normalizeBalanceRechargeMultiplier(cfg.BalanceRechargeMultiplier)
}

// ensureBackfilled folds the usage logged before this feature existed into the summary, once.
func (s *ProfitService) ensureBackfilled(ctx context.Context) error {
	s.backfill.Lock()
	defer s.backfill.Unlock()
	if s.backfilled {
		return nil
	}
	if s.settingRepo != nil {
		if v, err := s.settingRepo.GetValue(ctx, SettingKeyProfitBackfilled); err == nil && v != "" {
			s.backfilled = true
			return nil
		}
	}
	if err := s.repo.RebuildSummary(ctx); err != nil {
		return err
	}
	if s.settingRepo != nil {
		if err := s.settingRepo.Set(ctx, SettingKeyProfitBackfilled, s.now().UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	s.backfilled = true
	return nil
}

// Rebuild folds every usage log still kept into the summary again.
func (s *ProfitService) Rebuild(ctx context.Context) error {
	return s.repo.RebuildSummary(ctx)
}

func parseProfitDate(v string) (time.Time, bool) {
	t, err := time.Parse(profitDateLayout, strings.TrimSpace(v))
	return t, err == nil
}

func normalizeCostPackInput(in CostPackInput, creating bool) (CostPackInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Note = strings.TrimSpace(in.Note)
	if in.Name == "" || len([]rune(in.Name)) > costPackNameMaxRunes {
		return in, ErrCostPackInvalid
	}
	start, ok := parseProfitDate(in.StartedAt)
	if !ok {
		return in, ErrCostPackInvalid
	}
	in.StartedAt = start.Format(profitDateLayout)
	if in.EndedAt != nil && strings.TrimSpace(*in.EndedAt) == "" {
		in.EndedAt = nil
	}
	if in.EndedAt != nil {
		end, ok := parseProfitDate(*in.EndedAt)
		if !ok || end.Before(start) {
			return in, ErrCostPackInvalid
		}
		v := end.Format(profitDateLayout)
		in.EndedAt = &v
	}
	if in.CycleDays < 0 || in.CycleDays > profitMaxCycleDays {
		return in, ErrCostPackInvalid
	}
	if in.CycleDays > 0 {
		if !(in.CycleAmount > 0) || in.CycleAmount > profitMaxAmount {
			return in, ErrCostPackInvalid
		}
		in.Amount = 0
	} else {
		in.CycleAmount = 0
		if creating && (!(in.Amount > 0) || in.Amount > profitMaxAmount) {
			return in, ErrCostPackInvalid
		}
	}
	seen := make(map[int64]struct{}, len(in.AccountIDs))
	ids := make([]int64, 0, len(in.AccountIDs))
	for _, id := range in.AccountIDs {
		if id <= 0 {
			return in, ErrCostPackInvalid
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	in.AccountIDs = ids
	return in, nil
}

// costBetween is what the pack paid between two days (both included), never counting days after today.
func (p *CostPack) costBetween(from, to, today string) float64 {
	if to > today {
		to = today
	}
	if from > to {
		return 0
	}
	total := 0.0
	for _, pay := range p.Payments {
		if pay.PaidAt >= from && pay.PaidAt <= to {
			total += pay.Amount
		}
	}
	if p.CycleDays <= 0 {
		return total
	}
	last := to
	if p.EndedAt != nil && *p.EndedAt < last {
		last = *p.EndedAt
	}
	start, ok := parseProfitDate(p.StartedAt)
	if !ok {
		return total
	}
	for day := start; ; day = day.AddDate(0, 0, p.CycleDays) {
		d := day.Format(profitDateLayout)
		if d > last {
			break
		}
		if d >= from {
			total += p.CycleAmount
		}
	}
	return total
}

// nextCycle is the first cycle day after today, nil for one-off packs and packs that end before it.
func (p *CostPack) nextCycle(today string) *string {
	if p.CycleDays <= 0 {
		return nil
	}
	start, ok := parseProfitDate(p.StartedAt)
	if !ok {
		return nil
	}
	day := start
	for day.Format(profitDateLayout) <= today {
		day = day.AddDate(0, 0, p.CycleDays)
	}
	next := day.Format(profitDateLayout)
	if p.EndedAt != nil && next > *p.EndedAt {
		return nil
	}
	return &next
}

// ListPacks returns every pack with what it has cost and earned so far.
func (s *ProfitService) ListPacks(ctx context.Context) ([]CostPack, error) {
	if err := s.ensureBackfilled(ctx); err != nil {
		return nil, err
	}
	packs, err := s.repo.ListPacks(ctx)
	if err != nil {
		return nil, err
	}
	usage, err := s.repo.PackUsage(ctx, timezone.Name())
	if err != nil {
		return nil, err
	}
	today := s.today()
	rate := s.unitsPerYuan(ctx)
	for i := range packs {
		p := &packs[i]
		u := usage[p.ID]
		p.Paid = p.costBetween("", today, today)
		p.Income = u.Income / rate
		p.PlanUsage = u.PlanUsage / rate
		p.Requests = u.Requests
		p.Profit = p.Income - p.Paid
		p.Ended = p.EndedAt != nil && *p.EndedAt < today
		p.NextCycleAt = p.nextCycle(today)
	}
	return packs, nil
}

func (s *ProfitService) CreatePack(ctx context.Context, in CostPackInput) (int64, error) {
	in, err := normalizeCostPackInput(in, true)
	if err != nil {
		return 0, err
	}
	return s.repo.CreatePack(ctx, in)
}

func (s *ProfitService) UpdatePack(ctx context.Context, id int64, in CostPackInput) error {
	in, err := normalizeCostPackInput(in, false)
	if err != nil {
		return err
	}
	return s.repo.UpdatePack(ctx, id, in)
}

func (s *ProfitService) DeletePack(ctx context.Context, id int64) error {
	return s.repo.DeletePack(ctx, id)
}

func (s *ProfitService) AddPayment(ctx context.Context, packID int64, amount float64, paidAt, note string) error {
	day, ok := parseProfitDate(paidAt)
	if !ok || !(amount > 0) || amount > profitMaxAmount {
		return ErrCostPackInvalid
	}
	return s.repo.AddPayment(ctx, packID, amount, day.Format(profitDateLayout), strings.TrimSpace(note))
}

func (s *ProfitService) DeletePayment(ctx context.Context, packID, paymentID int64) error {
	return s.repo.DeletePayment(ctx, packID, paymentID)
}

// Report works out income, cost and profit between two days (both included).
func (s *ProfitService) Report(ctx context.Context, from, to string) (*ProfitReport, error) {
	start, ok := parseProfitDate(from)
	end, ok2 := parseProfitDate(to)
	if !ok || !ok2 || end.Before(start) || end.Sub(start) > profitMaxReportDays*24*time.Hour {
		return nil, ErrProfitRangeInvalid
	}
	from, to = start.Format(profitDateLayout), end.Format(profitDateLayout)
	if err := s.ensureBackfilled(ctx); err != nil {
		return nil, err
	}
	rows, err := s.repo.Usage(ctx, from, to, timezone.Name())
	if err != nil {
		return nil, err
	}
	packs, err := s.repo.ListPacks(ctx)
	if err != nil {
		return nil, err
	}
	rate := s.unitsPerYuan(ctx)
	report := buildProfitReport(rows, packs, from, to, s.today(), rate)

	accountIDs := make([]int64, 0, len(report.Accounts))
	for _, a := range report.Accounts {
		accountIDs = append(accountIDs, a.AccountID)
	}
	infos, err := s.repo.AccountInfos(ctx, accountIDs)
	if err != nil {
		return nil, err
	}
	for i := range report.Accounts {
		a := &report.Accounts[i]
		if info, ok := infos[a.AccountID]; ok {
			a.Name, a.Platform, a.Deleted, a.RateMultiplier = info.Name, info.Platform, info.Deleted, info.RateMultiplier
		}
	}
	groupIDs := make([]int64, 0, len(report.Groups))
	for _, g := range report.Groups {
		if g.GroupID > 0 {
			groupIDs = append(groupIDs, g.GroupID)
		}
	}
	names, err := s.repo.GroupNames(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	for i := range report.Groups {
		report.Groups[i].Name = names[report.Groups[i].GroupID]
	}
	return report, nil
}

// buildProfitReport turns summed usage and the packs into the report. rate is balance units per yuan.
func buildProfitReport(rows []ProfitUsageRow, packs []CostPack, from, to, today string, rate float64) *ProfitReport {
	if !(rate > 0) {
		rate = defaultBalanceRechargeMultiplier
	}
	// Empty lists, not nil: the page reads their length.
	report := &ProfitReport{
		From: from, To: to, UnitsPerYuan: rate,
		Months: []ProfitMonthLine{}, Groups: []ProfitGroupLine{}, Packs: []ProfitPackLine{}, Accounts: []ProfitAccountLine{},
	}
	months := map[string]*ProfitFigures{}
	groups := map[int64]*ProfitFigures{}
	packLines := map[int64]*ProfitFigures{}
	accounts := map[int64]*ProfitAccountLine{}
	// What each pack's accounts used on each group at standard price, to split the pack's cost between groups.
	packGroupUse := map[int64]map[int64]float64{}
	packNames := make(map[int64]string, len(packs))
	for i := range packs {
		packNames[packs[i].ID] = packs[i].Name
	}

	figures := func(m map[string]*ProfitFigures, key string) *ProfitFigures {
		if m[key] == nil {
			m[key] = &ProfitFigures{}
		}
		return m[key]
	}
	byID := func(m map[int64]*ProfitFigures, key int64) *ProfitFigures {
		if m[key] == nil {
			m[key] = &ProfitFigures{}
		}
		return m[key]
	}

	for _, row := range rows {
		var income, planUsage, rateCost float64
		if row.BillingType == BillingTypeSubscription {
			planUsage = row.ActualCost / rate
		} else {
			income = row.ActualCost / rate
		}
		if row.PackID == 0 {
			rateCost = row.AccountCost / rate
		} else {
			if packGroupUse[row.PackID] == nil {
				packGroupUse[row.PackID] = map[int64]float64{}
			}
			packGroupUse[row.PackID][row.GroupID] += row.StandardCost
		}
		add := func(f *ProfitFigures) {
			f.Requests += row.Requests
			f.Income += income
			f.PlanUsage += planUsage
			f.RateCost += rateCost
		}
		add(&report.Totals)
		add(figures(months, row.Month))
		add(byID(groups, row.GroupID))
		if row.PackID != 0 {
			add(byID(packLines, row.PackID))
		}
		acc := accounts[row.AccountID]
		if acc == nil {
			acc = &ProfitAccountLine{AccountID: row.AccountID}
			accounts[row.AccountID] = acc
		}
		add(&acc.ProfitFigures)
		if row.PackID != 0 {
			acc.PackID, acc.PackName = row.PackID, packNames[row.PackID]
		}
	}

	unallocated := 0.0
	for i := range packs {
		p := &packs[i]
		cost := p.costBetween(from, to, today)
		if cost != 0 || packLines[p.ID] != nil {
			byID(packLines, p.ID).PackCost = cost
		}
		if cost == 0 {
			continue
		}
		report.Totals.PackCost += cost
		used := 0.0
		for _, v := range packGroupUse[p.ID] {
			used += v
		}
		if used > 0 {
			for groupID, v := range packGroupUse[p.ID] {
				byID(groups, groupID).PackCost += cost * v / used
			}
		} else {
			unallocated += cost
		}
		for month := monthOf(from); month <= monthOf(to); month = nextMonth(month) {
			first, last := month+"-01", lastDayOfMonth(month)
			if c := p.costBetween(maxDate(from, first), minDate(to, last), today); c != 0 {
				figures(months, month).PackCost += c
			}
		}
	}
	if unallocated != 0 {
		byID(groups, profitUnallocatedID).PackCost += unallocated
	}
	report.Totals.settle()

	for month, f := range months {
		f.settle()
		report.Months = append(report.Months, ProfitMonthLine{Month: month, ProfitFigures: *f})
	}
	sort.Slice(report.Months, func(i, j int) bool { return report.Months[i].Month < report.Months[j].Month })

	for id, f := range groups {
		f.settle()
		report.Groups = append(report.Groups, ProfitGroupLine{GroupID: id, ProfitFigures: *f})
	}
	sort.Slice(report.Groups, func(i, j int) bool {
		a, b := report.Groups[i], report.Groups[j]
		if (a.GroupID == profitUnallocatedID) != (b.GroupID == profitUnallocatedID) {
			return b.GroupID == profitUnallocatedID
		}
		if a.Income != b.Income {
			return a.Income > b.Income
		}
		return a.GroupID < b.GroupID
	})

	for id, f := range packLines {
		f.settle()
		report.Packs = append(report.Packs, ProfitPackLine{PackID: id, Name: packNames[id], ProfitFigures: *f})
	}
	sort.Slice(report.Packs, func(i, j int) bool {
		if report.Packs[i].Profit != report.Packs[j].Profit {
			return report.Packs[i].Profit < report.Packs[j].Profit
		}
		return report.Packs[i].PackID < report.Packs[j].PackID
	})

	for _, acc := range accounts {
		acc.settle()
		report.Accounts = append(report.Accounts, *acc)
	}
	sort.Slice(report.Accounts, func(i, j int) bool {
		a, b := report.Accounts[i], report.Accounts[j]
		if a.Income != b.Income {
			return a.Income > b.Income
		}
		return a.AccountID < b.AccountID
	})
	return report
}

func maxDate(a, b string) string {
	if a > b {
		return a
	}
	return b
}

func minDate(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func monthOf(day string) string {
	if len(day) < 7 {
		return day
	}
	return day[:7]
}

func nextMonth(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "9999-12"
	}
	return t.AddDate(0, 1, 0).Format("2006-01")
}

func lastDayOfMonth(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month + "-31"
	}
	return t.AddDate(0, 1, -1).Format(profitDateLayout)
}
