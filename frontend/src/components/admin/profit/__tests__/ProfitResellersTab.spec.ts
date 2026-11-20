import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ProfitResellersTab from '../ProfitResellersTab.vue'

const { list, add, remove, getPortal, setPortal, lookupUsers, showError, showSuccess } = vi.hoisted(() => ({
  getPortal: vi.fn(),
  setPortal: vi.fn(),
  showSuccess: vi.fn(),
  list: vi.fn(),
  add: vi.fn(),
  remove: vi.fn(),
  lookupUsers: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin/resellers', () => ({ list, add, remove, getPortal, setPortal }))
vi.mock('@/api/admin/affiliates', () => ({ lookupUsers }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('@/composables/useYcMoney', () => ({ useYcMoney: () => ({ yuan: (units: number) => `¥${(units / 2).toFixed(2)}` }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key) })
}))

const reseller = { user_id: 7, email: 'shop@example.com', since: '2026-11-01T00:00:00Z', balance: 200, keys: 2, client_used: 30, cost: 20 }

const mountTab = async () => {
  const wrapper = mount(ProfitResellersTab)
  await flushPromises()
  return wrapper
}

describe('ProfitResellersTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    list.mockResolvedValue([reseller])
    add.mockResolvedValue(undefined)
    remove.mockResolvedValue(undefined)
    getPortal.mockResolvedValue({ domain: 'keys.example.com', name: 'Keys', contact: '', logo: '' })
    setPortal.mockImplementation(async (portal) => portal)
    lookupUsers.mockResolvedValue([{ id: 12, email: 'new@example.com', username: 'new' }])
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('lists resellers with their balance, keys and what their clients used', async () => {
    const wrapper = await mountTab()
    const row = wrapper.get('[data-test="resellers-row"]').text()
    expect(row).toContain('shop@example.com')
    expect(row).toContain('¥100.00')
    expect(row).toContain('¥15.00')
    expect(row).toContain('¥10.00')
  })

  it('adds a searched user', async () => {
    const wrapper = await mountTab()
    expect(wrapper.get('[data-test="resellers-add"]').attributes('disabled')).toBeDefined()
    await wrapper.get('[data-test="resellers-search"]').setValue('new')
    await flushPromises()
    await wrapper.get('[data-test="resellers-option"]').trigger('click')
    await wrapper.get('[data-test="resellers-add"]').trigger('click')
    await flushPromises()
    expect(add).toHaveBeenCalledWith(12)
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('removes a reseller only after a yes', async () => {
    const wrapper = await mountTab()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    await wrapper.get('[data-test="resellers-remove"]').trigger('click')
    expect(remove).not.toHaveBeenCalled()

    vi.spyOn(window, 'confirm').mockReturnValue(true)
    await wrapper.get('[data-test="resellers-remove"]').trigger('click')
    await flushPromises()
    expect(remove).toHaveBeenCalledWith(7)
  })

  it('loads and saves the client domain, and wants a name with a domain', async () => {
    const wrapper = await mountTab()
    expect((wrapper.get('[data-test="portal-domain"]').element as HTMLInputElement).value).toBe('keys.example.com')

    await wrapper.get('[data-test="portal-name"]').setValue('')
    await wrapper.get('[data-test="resellers-portal"] form').trigger('submit')
    expect(setPortal).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('admin.profit.resellers.portal.needName')

    await wrapper.get('[data-test="portal-name"]').setValue('Client keys')
    await wrapper.get('[data-test="portal-contact"]').setValue('wx: keys')
    await wrapper.get('[data-test="resellers-portal"] form').trigger('submit')
    await flushPromises()
    expect(setPortal).toHaveBeenCalledWith({ domain: 'keys.example.com', name: 'Client keys', contact: 'wx: keys', logo: '' })
    expect(showSuccess).toHaveBeenCalledWith('admin.profit.resellers.portal.saved')
  })
})
