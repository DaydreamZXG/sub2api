import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AgentUsersDialog from '../AgentUsersDialog.vue'

const { listUsers, attachUser, detachUser, lookupUsers } = vi.hoisted(() => ({
  listUsers: vi.fn(),
  attachUser: vi.fn(),
  detachUser: vi.fn(),
  lookupUsers: vi.fn()
}))

vi.mock('@/api/admin/agents', () => ({ listUsers, attachUser, detachUser }))
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

const agent = { user_id: 7, email: 'agent@example.com', share_percent: 10, since: null, invitees: 1, shared: 84, available: 60 }
const BaseDialog = { props: ['show'], template: '<div v-if="show"><slot /></div>' }
const mountDialog = async () => {
  const wrapper = mount(AgentUsersDialog, { props: { agent }, global: { stubs: { BaseDialog } } })
  await flushPromises()
  return wrapper
}

describe('AgentUsersDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listUsers.mockResolvedValue([
      { user_id: 2, email: 'acme@example.com', joined_at: '2026-11-19T02:00:00Z', attached: true, spent: 80, share: 8 }
    ])
    attachUser.mockResolvedValue(undefined)
    detachUser.mockResolvedValue(undefined)
    lookupUsers.mockResolvedValue([{ id: 12, email: 'new@example.com', username: 'new' }])
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('lists the users under the agent with amounts in yuan', async () => {
    const wrapper = await mountDialog()
    expect(listUsers).toHaveBeenCalledWith(7)
    const row = wrapper.get('[data-test="agent-users-row"]').text()
    expect(row).toContain('acme@example.com')
    expect(row).toContain('admin.profit.agents.users.attachedAt')
    expect(row).toContain('"spent":"¥40.00","share":"¥4.00"')
  })

  it('puts a searched user under the agent after a confirmation', async () => {
    const wrapper = await mountDialog()
    await wrapper.get('[data-test="agent-users-search"]').setValue('new')
    await flushPromises()
    await wrapper.get('[data-test="agent-users-option"]').trigger('click')
    await wrapper.get('[data-test="agent-users-add"]').trigger('click')
    await flushPromises()
    expect(attachUser).toHaveBeenCalledWith(7, 12)
    expect(listUsers).toHaveBeenCalledTimes(2)
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })

  it('does nothing when the confirmation is refused', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = await mountDialog()
    await wrapper.get('[data-test="agent-users-remove"]').trigger('click')
    await flushPromises()
    expect(detachUser).not.toHaveBeenCalled()
  })

  it('takes a user off the agent', async () => {
    const wrapper = await mountDialog()
    await wrapper.get('[data-test="agent-users-remove"]').trigger('click')
    await flushPromises()
    expect(detachUser).toHaveBeenCalledWith(7, 2)
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })
})
