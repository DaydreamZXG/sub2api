import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ProfitView from '../ProfitView.vue'

const { getReport, rebuild, listPacks, createPack, updatePack, deletePack, addPayment, deletePayment, getSettings, updateSettings, listAccounts } = vi.hoisted(
  () => ({
    getReport: vi.fn(),
    rebuild: vi.fn(),
    listPacks: vi.fn(),
    createPack: vi.fn(),
    updatePack: vi.fn(),
    deletePack: vi.fn(),
    addPayment: vi.fn(),
    deletePayment: vi.fn(),
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    listAccounts: vi.fn()
  })
)

vi.mock('@/api/admin/profit', () => ({
  getReport, rebuild, listPacks, createPack, updatePack, deletePack, addPayment, deletePayment, getSettings, updateSettings
}))
vi.mock('@/api/admin/accounts', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/api/admin/accounts')>()),
  list: listAccounts
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key)
  })
}))

const figures = (over: Record<string, number> = {}) => ({
  requests: 0, income: 0, gifted: 0, plan_usage: 0, rate_cost: 0, pack_cost: 0, profit: 0, ...over
})

const report = {
  from: '2026-11-01',
  to: '2026-11-20',
  units_per_yuan: 1,
  received: 900,
  refunded: 50,
  totals: figures({ requests: 57, income: 1520.5, gifted: 210, plan_usage: 25, rate_cost: 150, pack_cost: 346, profit: 1024.5 }),
  months: [],
  groups: [
    { group_id: 1, name: 'Claude', ...figures({ income: 1200, pack_cost: 336, profit: 864 }) },
    { group_id: -1, name: '', ...figures({ pack_cost: 10, profit: -10 }) }
  ],
  packs: [{ pack_id: 3, name: 'Idle', ...figures({ pack_cost: 10, profit: -10 }) }],
  accounts: [
    { account_id: 10, name: 'max-1', platform: 'anthropic', deleted: false, pack_id: 2, pack_name: 'Max', rate_multiplier: 1, ...figures({ requests: 40, income: 800 }) },
    { account_id: 30, name: 'cheap', platform: 'openai', deleted: false, pack_id: 0, pack_name: '', rate_multiplier: 0.3, ...figures({ requests: 17, income: 720.5, rate_cost: 150 }) }
  ]
}

const pack = (over: Record<string, unknown> = {}) => ({
  id: 2, name: 'Max', note: '', cycle_days: 30, cycle_amount: 1400, started_at: '2026-11-01', ended_at: null,
  payments: [], accounts: [{ id: 10, name: 'max-1', platform: 'anthropic', deleted: false }],
  paid: 1400, income: 868, plan_usage: 0, profit: -532, requests: 40, ended: false, next_cycle_at: '2026-12-01',
  ...over
})

function mountView() {
  return mount(ProfitView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: { props: ['show', 'title'], template: '<div v-if="show"><slot /></div>' }
      }
    }
  })
}

describe('ProfitView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getReport.mockResolvedValue(report)
    listPacks.mockResolvedValue([pack(), pack({ id: 5, name: 'Top-up', cycle_days: 0, cycle_amount: 0, paid: 36, income: 80, profit: 44, next_cycle_at: null })])
    createPack.mockResolvedValue(undefined)
    getSettings.mockResolvedValue({ manual_is_paid: false, redeem_is_paid: true })
    updateSettings.mockImplementation(async (value: unknown) => value)
    listAccounts.mockResolvedValue({ items: [], total: 0 })
  })

  it('shows the totals in yuan and names the pack cost nobody used', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(getReport).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-test="profit-tile-income"]').text()).toContain('¥1,520.50')
    expect(wrapper.get('[data-test="profit-tile-profit"]').text()).toContain('¥1,024.50')
    expect(wrapper.get('[data-test="profit-tile-margin"]').text()).toContain('67.4%')
    expect(wrapper.get('[data-test="profit-plan-note"]').text()).toContain('¥25.00')
    expect(wrapper.get('[data-test="profit-gifted-note"]').text()).toContain('¥210.00')
    const cash = wrapper.get('[data-test="profit-cash-note"]').text()
    expect(cash).toContain('¥900.00')
    expect(cash).toContain('¥50.00')

    const groups = wrapper.get('[data-test="profit-by-group"]').text()
    expect(groups).toContain('admin.profit.unallocated')
    expect(groups).toContain('-¥10.00')

    const accounts = wrapper.get('[data-test="profit-by-account"]').text()
    expect(accounts).toContain('admin.profit.costByPack:{"name":"Max"}')
    expect(accounts).toContain('admin.profit.costByRate:{"rate":0.3}')
    // One month only: the monthly table would repeat the totals.
    expect(wrapper.find('[data-test="profit-by-month"]').exists()).toBe(false)
  })

  it('says how far each pack is from paying for itself', async () => {
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-test="profit-tab-packs"]').trigger('click')

    const cards = wrapper.findAll('[data-test="pack-card"]')
    expect(cards).toHaveLength(2)
    expect(cards[0].get('[data-test="pack-payback"]').text()).toBe('admin.profit.paybackSoFar:{"percent":62,"amount":"¥532.00"}')
    expect(cards[0].get('[data-test="pack-profit"]').text()).toBe('-¥532.00')
    expect(cards[1].get('[data-test="pack-payback"]').text()).toBe('admin.profit.paybackDone:{"amount":"¥44.00"}')
  })

  it('asks for a name and an amount before creating a pack', async () => {
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-test="profit-tab-packs"]').trigger('click')
    await wrapper.get('[data-test="pack-new"]').trigger('click')

    await wrapper.get('[data-test="pack-form"]').trigger('submit')
    expect(wrapper.get('[data-test="pack-form-error"]').text()).toBe('admin.profit.form.errors.name')
    await wrapper.get('[data-test="pack-name"]').setValue('  Pay 5 get 100 ')
    await wrapper.get('[data-test="pack-form"]').trigger('submit')
    expect(wrapper.get('[data-test="pack-form-error"]').text()).toBe('admin.profit.form.errors.amount')
    expect(createPack).not.toHaveBeenCalled()

    await wrapper.get('[data-test="pack-amount"]').setValue('36')
    await wrapper.get('[data-test="pack-form"]').trigger('submit')
    await flushPromises()

    expect(createPack).toHaveBeenCalledTimes(1)
    expect(createPack.mock.calls[0][0]).toMatchObject({
      name: 'Pay 5 get 100', cycle_days: 0, cycle_amount: 0, amount: 36, ended_at: null, account_ids: []
    })
    expect(wrapper.find('[data-test="pack-form"]').exists()).toBe(false)
  })

  it('sends a subscription as a fee per period without a one-off price', async () => {
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-test="profit-tab-packs"]').trigger('click')
    await wrapper.get('[data-test="pack-new"]').trigger('click')
    await wrapper.get('[data-test="pack-name"]').setValue('Max')
    await wrapper.get('[data-test="pack-kind-cycle"]').trigger('click')
    await wrapper.get('[data-test="pack-cycle-amount"]').setValue('1400')
    await wrapper.get('[data-test="pack-form"]').trigger('submit')
    await flushPromises()

    expect(createPack.mock.calls[0][0]).toMatchObject({ name: 'Max', cycle_days: 30, cycle_amount: 1400, amount: 0 })
  })

  it('saves both income switches together and loads the report again', async () => {
    const wrapper = mountView()
    await flushPromises()

    const manual = wrapper.get('[data-test="profit-basis-manual"]')
    expect(manual.attributes('aria-checked')).toBe('false')
    expect(wrapper.get('[data-test="profit-basis-redeem"]').attributes('aria-checked')).toBe('true')

    await manual.trigger('click')
    await flushPromises()

    expect(updateSettings).toHaveBeenCalledWith({ manual_is_paid: true, redeem_is_paid: true })
    expect(wrapper.get('[data-test="profit-basis-manual"]').attributes('aria-checked')).toBe('true')
    expect(getReport).toHaveBeenCalledTimes(2)
  })
})
