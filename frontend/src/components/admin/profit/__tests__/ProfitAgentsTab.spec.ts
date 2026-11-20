import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ProfitAgentsTab from '../ProfitAgentsTab.vue'

const { list, set, remove, listPayouts, markPaid, reject, lookupUsers } = vi.hoisted(() => ({
  list: vi.fn(),
  set: vi.fn(),
  remove: vi.fn(),
  listPayouts: vi.fn(),
  markPaid: vi.fn(),
  reject: vi.fn(),
  lookupUsers: vi.fn()
}))

vi.mock('@/api/admin/agents', () => ({ list, set, remove, listPayouts, markPaid, reject }))
vi.mock('@/api/admin/affiliates', () => ({ lookupUsers }))
vi.mock('@/composables/useYcMoney', () => ({
  useYcMoney: () => ({
    yuan: (units: number) => `¥${(units / 2).toFixed(2)}`,
    toYuan: (units: number) => units / 2,
    toUnits: (yuanAmount: number) => yuanAmount * 2
  })
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key)
  })
}))

const agent = { user_id: 7, email: 'agent@example.com', share_percent: 10, since: '2026-11-01T00:00:00Z', invitees: 3, shared: 84, available: 60 }
const payout = (over: Record<string, unknown> = {}) => ({
  id: 5, user_id: 7, email: 'agent@example.com', amount: 40, payee_note: 'Alipay 138', status: 'pending', admin_note: '',
  created_at: '2026-11-19T02:00:00Z', handled_at: null, ...over
})

const BaseDialog = { props: ['show'], template: '<div v-if="show"><slot /></div>' }
const mountTab = async () => {
  const wrapper = mount(ProfitAgentsTab, { global: { stubs: { BaseDialog } } })
  await flushPromises()
  return wrapper
}

describe('ProfitAgentsTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    list.mockResolvedValue([agent])
    listPayouts.mockResolvedValue([payout(), payout({ id: 4, status: 'paid', admin_note: 'done', handled_at: '2026-11-10T02:00:00Z' })])
    set.mockResolvedValue(undefined)
    markPaid.mockResolvedValue(undefined)
    reject.mockResolvedValue(undefined)
    lookupUsers.mockResolvedValue([{ id: 12, email: 'new@example.com', username: 'new' }])
  })

  it('lists agents and requests with amounts in yuan', async () => {
    const wrapper = await mountTab()
    const row = wrapper.find('[data-test="agents-row"]').text()
    expect(row).toContain('agent@example.com')
    expect(row).toContain('10%')
    expect(row).toContain('¥42.00')
    expect(row).toContain('¥30.00')
    expect(wrapper.findAll('[data-test="agents-pending-row"]')).toHaveLength(1)
    expect(wrapper.find('[data-test="agents-pending-row"]').text()).toContain('¥20.00')
    expect(wrapper.find('[data-test="agents-handled"]').text()).toContain('done')
  })

  it('marks a request paid with the note and reloads', async () => {
    const wrapper = await mountTab()
    await wrapper.find('[data-test="agents-pay"]').trigger('click')
    await wrapper.find('[data-test="agents-note"]').setValue('sent by Alipay')
    await wrapper.find('[data-test="agents-decide"]').trigger('submit')
    await flushPromises()
    expect(markPaid).toHaveBeenCalledWith(5, 'sent by Alipay')
    expect(reject).not.toHaveBeenCalled()
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('rejects a request', async () => {
    const wrapper = await mountTab()
    await wrapper.find('[data-test="agents-reject"]').trigger('click')
    await wrapper.find('[data-test="agents-decide"]').trigger('submit')
    await flushPromises()
    expect(reject).toHaveBeenCalledWith(5, '')
    expect(markPaid).not.toHaveBeenCalled()
  })

  it('makes a searched user an agent at the entered share', async () => {
    const wrapper = await mountTab()
    await wrapper.find('[data-test="agents-add"]').trigger('click')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    expect(set).not.toHaveBeenCalled()

    await wrapper.find('[data-test="agents-user-search"]').setValue('new')
    await flushPromises()
    await wrapper.find('[data-test="agents-user-option"]').trigger('click')
    await wrapper.find('[data-test="agents-share"]').setValue('12.5')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    await flushPromises()
    expect(lookupUsers).toHaveBeenCalledWith('new')
    expect(set).toHaveBeenCalledWith(12, 12.5, [])
  })

  it('shows the share in force for an agent with tiers', async () => {
    list.mockResolvedValue([
      { ...agent, share_percent: 5, tiers: [{ from: 10000, percent: 8 }, { from: 40000, percent: 10 }], month_spent: 12000, current_percent: 8 }
    ])
    const wrapper = await mountTab()
    const cell = wrapper.get('[data-test="agents-share-cell"]').text()
    expect(cell).toContain('8%')
    expect(cell).toContain('"from":5,"to":10,"spent":"¥6000.00"')
  })

  it('edits tiers in yuan and sends them in balance units', async () => {
    list.mockResolvedValue([{ ...agent, share_percent: 5, tiers: [{ from: 10000, percent: 8 }], month_spent: 0, current_percent: 5 }])
    const wrapper = await mountTab()
    await wrapper.find('[data-test="agents-row"] button:nth-of-type(2)').trigger('click')
    // The stored 10000 units are shown as the ¥5000 they stand for.
    expect((wrapper.get('[data-test="agents-tier-from"]').element as HTMLInputElement).value).toBe('5000')

    await wrapper.get('[data-test="agents-tier-add"]').trigger('click')
    const rows = wrapper.findAll('[data-test="agents-tier-row"]')
    expect(rows).toHaveLength(2)
    await rows[1].get('[data-test="agents-tier-from"]').setValue('20000')
    await rows[1].get('[data-test="agents-tier-percent"]').setValue('7')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    // 7% is not above the 8% of the tier before it.
    expect(set).not.toHaveBeenCalled()

    await rows[1].get('[data-test="agents-tier-percent"]').setValue('10')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    await flushPromises()
    expect(set).toHaveBeenCalledWith(7, 5, [{ from: 10000, percent: 8 }, { from: 40000, percent: 10 }])
  })

  it('sends back thresholds that were not retyped exactly as they were stored', async () => {
    // 2 balance units to the yuan: 100.01 and 100.02 units both show as ¥50.01.
    list.mockResolvedValue([
      { ...agent, share_percent: 5, tiers: [{ from: 100.01, percent: 8 }, { from: 100.02, percent: 9 }], month_spent: 0, current_percent: 5 }
    ])
    const wrapper = await mountTab()
    await wrapper.find('[data-test="agents-row"] button:nth-of-type(2)').trigger('click')
    await wrapper.find('[data-test="agents-share"]').setValue('6')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    await flushPromises()
    expect(set).toHaveBeenCalledWith(7, 6, [{ from: 100.01, percent: 8 }, { from: 100.02, percent: 9 }])

    // A retyped amount is sent as typed.
    await wrapper.find('[data-test="agents-row"] button:nth-of-type(2)').trigger('click')
    const rows = wrapper.findAll('[data-test="agents-tier-row"]')
    await rows[1].get('[data-test="agents-tier-from"]').setValue('60')
    await wrapper.find('[data-test="agents-save"]').trigger('submit')
    await flushPromises()
    expect(set).toHaveBeenLastCalledWith(7, 5, [{ from: 100.01, percent: 8 }, { from: 120, percent: 9 }])
  })
})
