import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ResellerView from '../ResellerView.vue'

const { getReseller, createClientKey, getAvailable, toggleStatus, deleteKey, showError, copyToClipboard } = vi.hoisted(() => ({
  getReseller: vi.fn(),
  createClientKey: vi.fn(),
  getAvailable: vi.fn(),
  toggleStatus: vi.fn(),
  deleteKey: vi.fn(),
  showError: vi.fn(),
  copyToClipboard: vi.fn()
}))

vi.mock('@/api/reseller', () => ({ default: { getReseller, createClientKey } }))
vi.mock('@/api/groups', () => ({ default: { getAvailable } }))
vi.mock('@/api/keys', () => ({ default: { toggleStatus, delete: deleteKey } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard }) }))
vi.mock('@/composables/ycMoneyState', () => ({ ycSign: { value: '¥' } }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key) })
}))

const key = (over: Record<string, unknown> = {}) => ({
  id: 8, name: 'studio', key: 'sk-client', group_id: 3, group_name: 'Claude', status: 'active', markup: 1.5,
  budget: 10, client_used: 3, cost: 2, margin: 1, expires_at: null, last_used_at: null, created_at: '2026-11-20T00:00:00Z', ...over
})
const overview = (over: Record<string, unknown> = {}) => ({
  balance: 125, client_used: 3, cost: 2, margin: 1, reserved: 4.67, keys: [key()], ...over
})

const stubs = {
  AppLayout: { template: '<main><slot /></main>' },
  RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }
}
const mountView = async () => {
  const wrapper = mount(ResellerView, { global: { stubs } })
  await flushPromises()
  return wrapper
}

describe('ResellerView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getReseller.mockResolvedValue(overview())
    getAvailable.mockResolvedValue([{ id: 3, name: 'Claude' }])
    createClientKey.mockResolvedValue({ id: 9, key: 'sk-new' })
    toggleStatus.mockResolvedValue({})
    deleteKey.mockResolvedValue({})
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('tells a user who is not a reseller how to become one, and asks for nothing else', async () => {
    getReseller.mockResolvedValue(null)
    const wrapper = await mountView()
    expect(wrapper.find('[data-test="reseller-none"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="reseller-form"]').exists()).toBe(false)
    expect(getAvailable).not.toHaveBeenCalled()
  })

  it('shows what clients used, what it cost and the margin, per key and in all', async () => {
    const wrapper = await mountView()
    expect(wrapper.get('[data-test="reseller-balance"]').text()).toBe('¥125.00')
    expect(wrapper.get('[data-test="reseller-margin"]').text()).toBe('¥1.00')
    const row = wrapper.get('[data-test="reseller-key"]').text()
    expect(row).toContain('"used":"¥3.00","budget":"¥10.00"')
    expect(row).toContain('"cost":"¥2.00","margin":"¥1.00"')
    expect(row).toContain('"markup":1.5')
    expect(wrapper.find('[data-test="reseller-short"]').exists()).toBe(false)
  })

  it('warns when the balance cannot cover what clients have left', async () => {
    getReseller.mockResolvedValue(overview({ balance: 2 }))
    const wrapper = await mountView()
    expect(wrapper.get('[data-test="reseller-short"]').text()).toContain('"balance":"¥2.00","need":"¥4.67"')
  })

  it('says what a markup means before the key is made, and creates it', async () => {
    const wrapper = await mountView()
    await wrapper.get('[data-test="reseller-create"]').trigger('submit')
    expect(createClientKey).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('reseller.form.invalid')

    await wrapper.get('[data-test="reseller-name"]').setValue('shop')
    // The first line is already chosen: a key on no line could not make calls.
    expect((wrapper.get('[data-test="reseller-line"]').element as HTMLSelectElement).value).toBe('3')
    await wrapper.get('[data-test="reseller-budget"]').setValue('30')
    await wrapper.get('[data-test="reseller-markup"]').setValue('2')
    await wrapper.get('[data-test="reseller-days"]').setValue('30')
    expect(wrapper.get('[data-test="reseller-explain"]').text()).toContain('"client":"¥2.00","margin":"¥1.00","budget":"¥30.00","cost":"¥15.00"')

    await wrapper.get('[data-test="reseller-form"]').trigger('submit')
    await flushPromises()
    expect(createClientKey).toHaveBeenCalledWith({ name: 'shop', group_id: 3, budget: 30, markup: 2, expires_in_days: 30 })
    expect(copyToClipboard).toHaveBeenCalledWith('sk-new', 'reseller.form.created')
    expect(getReseller).toHaveBeenCalledTimes(2)
  })

  it('refuses a markup below cost', async () => {
    const wrapper = await mountView()
    await wrapper.get('[data-test="reseller-name"]').setValue('shop')
    await wrapper.get('[data-test="reseller-markup"]').setValue('0.5')
    expect(wrapper.get('[data-test="reseller-explain"]').text()).toBe('reseller.form.markupInvalid')
    await wrapper.get('[data-test="reseller-form"]').trigger('submit')
    expect(createClientKey).not.toHaveBeenCalled()
  })

  it('stops, starts and deletes a client key through the ordinary key API', async () => {
    const wrapper = await mountView()
    await wrapper.get('[data-test="reseller-toggle"]').trigger('click')
    await flushPromises()
    expect(toggleStatus).toHaveBeenCalledWith(8, 'inactive')

    await wrapper.get('[data-test="reseller-delete"]').trigger('click')
    await flushPromises()
    expect(deleteKey).toHaveBeenCalledWith(8)

    vi.spyOn(window, 'confirm').mockReturnValue(false)
    await wrapper.get('[data-test="reseller-delete"]').trigger('click')
    expect(deleteKey).toHaveBeenCalledTimes(1)
  })

  it('does not let a key be made when the reseller has no line', async () => {
    getAvailable.mockResolvedValue([])
    const wrapper = await mountView()
    expect(wrapper.find('[data-test="reseller-no-lines"]').exists()).toBe(true)
    expect(wrapper.get('[data-test="reseller-create"]').attributes('disabled')).toBeDefined()
    await wrapper.get('[data-test="reseller-name"]').setValue('shop')
    await wrapper.get('[data-test="reseller-form"]').trigger('submit')
    expect(createClientKey).not.toHaveBeenCalled()
  })
})
