import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AgentPanel from '../AgentPanel.vue'

const { requestAgentPayout, showError } = vi.hoisted(() => ({ requestAgentPayout: vi.fn(), showError: vi.fn() }))

vi.mock('@/api/user', () => ({ default: { requestAgentPayout } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key)
  })
}))

const agent = (over: Record<string, unknown> = {}) => ({
  share_percent: 10, since: '2026-11-01T00:00:00Z', available: 60.129, shared: 84, spent: 840, month_shared: 12.5, month_spent: 125,
  invitees: [{ user_id: 2, email: 'a***@example.com', joined_at: '2026-11-02T00:00:00Z', spent: 840, share: 84 }],
  payouts: [] as unknown[],
  ...over
})
const money = (value: number | null | undefined) => `¥${(value ?? 0).toFixed(2)}`
const mountPanel = (over: Record<string, unknown> = {}) => mount(AgentPanel, { props: { agent: agent(over) as never, money } })

describe('AgentPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    requestAgentPayout.mockResolvedValue(undefined)
  })

  it('shows the share, this month and the users behind it', () => {
    const wrapper = mountPanel()
    expect(wrapper.text()).toContain('affiliate.agent.rule:{"rate":"10%"}')
    expect(wrapper.find('[data-test="agent-month"]').text()).toBe('¥12.50')
    const row = wrapper.find('[data-test="agent-user-row"]').text()
    expect(row).toContain('a***@example.com')
    expect(row).toContain('¥84.00')
  })

  it('asks for a payout of what is available, with how to pay', async () => {
    const wrapper = mountPanel()
    await wrapper.find('[data-test="agent-payout-open"]').trigger('click')
    const amount = wrapper.find('[data-test="agent-payout-amount"]')
    expect((amount.element as HTMLInputElement).value).toBe('60.12')

    await wrapper.find('[data-test="agent-payout-form"]').trigger('submit')
    expect(requestAgentPayout).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('affiliate.agent.payout.payeeRequired')

    await wrapper.find('[data-test="agent-payout-payee"]').setValue(' Alipay 138 ')
    await wrapper.find('[data-test="agent-payout-form"]').trigger('submit')
    await flushPromises()
    expect(requestAgentPayout).toHaveBeenCalledWith(60.12, 'Alipay 138')
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })

  it('refuses more than what is available', async () => {
    const wrapper = mountPanel()
    await wrapper.find('[data-test="agent-payout-open"]').trigger('click')
    await wrapper.find('[data-test="agent-payout-amount"]').setValue('61')
    await wrapper.find('[data-test="agent-payout-payee"]').setValue('Alipay')
    await wrapper.find('[data-test="agent-payout-form"]').trigger('submit')
    expect(requestAgentPayout).not.toHaveBeenCalled()
  })

  it('shows a waiting request instead of the button', () => {
    const wrapper = mountPanel({
      payouts: [{ id: 1, amount: 40, payee_note: 'x', status: 'pending', admin_note: '', created_at: '2026-11-19T02:00:00Z', handled_at: null }]
    })
    expect(wrapper.find('[data-test="agent-payout-open"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="agent-pending"]').text()).toContain('¥40.00')
  })

  it('tells an agent with tiers how far the next one is', () => {
    const wrapper = mountPanel({
      share_percent: 5, tiers: [{ from: 500, percent: 8 }], current_percent: 5, next_tier: { from: 500, percent: 8 }, month_spent: 125
    })
    expect(wrapper.text()).toContain('"rate":"5%"')
    const tier = wrapper.get('[data-test="agent-tier"]').text()
    expect(tier).toContain('affiliate.agent.nextTier')
    expect(tier).toContain('"rate":"8%"')

    const top = mountPanel({ share_percent: 5, tiers: [{ from: 100, percent: 8 }], current_percent: 8, next_tier: null })
    expect(top.text()).toContain('"rate":"8%"')
    expect(top.get('[data-test="agent-tier"]').text()).toContain('affiliate.agent.topTier')
    expect(mountPanel().find('[data-test="agent-tier"]').exists()).toBe(false)
  })
})
