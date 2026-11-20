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
  useYcMoney: () => ({ yuan: (units: number) => `¥${(units / 2).toFixed(2)}` })
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
    expect(set).toHaveBeenCalledWith(12, 12.5)
  })
})
