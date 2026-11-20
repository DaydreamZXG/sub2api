import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PortalApp from '../PortalApp.vue'

const config = { portal: true as const, site_name: 'Keys', portal_logo: '', contact_info: 'wx: keys', sign: '¥' }
const usage = {
  status: 'active',
  quota: { limit: 100, used: 36, remaining: 64 },
  expires_at: '2026-12-20T00:00:00Z',
  days_until_expiry: 30,
  usage: { today: { requests: 2, total_tokens: 1500, actual_cost: 6 }, total: { requests: 9, total_tokens: 9000, actual_cost: 36 } },
  daily_usage: [
    { date: '2026-11-19', requests: 7, actual_cost: 30 },
    { date: '2026-11-20', requests: 2, actual_cost: 6 },
    { date: '2026-11-21', requests: 0, actual_cost: 0 }
  ],
  model_stats: [
    { model: 'small', requests: 6, actual_cost: 6 },
    { model: 'large', requests: 3, actual_cost: 30 }
  ]
}

const json = (body: unknown, status = 200) => ({ ok: status >= 200 && status < 300, status, json: async () => body })
const fetchMock = vi.fn()

const mountApp = async () => {
  const wrapper = mount(PortalApp, { props: { config } })
  await flushPromises()
  return wrapper
}

describe('PortalApp', () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem('portal_locale', 'en')
    fetchMock.mockReset()
    fetchMock.mockImplementation(async (url: string) =>
      url.startsWith('/v1/models') ? json({ data: [{ id: 'small' }, { id: 'large' }] }) : json(usage)
    )
    vi.stubGlobal('fetch', fetchMock)
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('asks for the key first and calls nothing before it has one', async () => {
    const wrapper = await mountApp()
    expect(wrapper.find('[data-test="portal-enter"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Keys')
    expect(wrapper.get('[data-test="portal-contact"]').text()).toContain('wx: keys')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('shows the budget, the usage and how to connect, asking only the gateway with the key', async () => {
    const wrapper = await mountApp()
    await wrapper.get('[data-test="portal-key-input"]').setValue('sk-0123456789abcdef')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual(['/v1/usage?days=30', '/v1/models'])
    expect(fetchMock.mock.calls[0][1]).toEqual({ headers: { Authorization: 'Bearer sk-0123456789abcdef' } })
    expect(wrapper.get('[data-test="portal-remaining"]').text()).toBe('¥64.00')
    expect(wrapper.get('[data-test="portal-status"]').text()).toBe('Active')
    expect(wrapper.get('[data-test="portal-budget"]').text()).toContain('¥36.00 used of ¥100.00')
    expect(wrapper.get('[data-test="portal-today"]').text()).toBe('¥6.00')
    expect(wrapper.get('[data-test="portal-total"]').text()).toBe('¥36.00')
    // Newest day first, days without calls left out; the dearest model first.
    const days = wrapper.findAll('[data-test="portal-daily"] tbody tr').map((row) => row.text())
    expect(days).toHaveLength(2)
    expect(days[0]).toContain('2026-11-20')
    expect(wrapper.findAll('[data-test="portal-by-model"] tbody tr')[0].text()).toContain('large')
    expect(wrapper.get('[data-test="portal-models"]').text()).toContain('small')
    expect(wrapper.get('[data-test="portal-base-url"]').text()).toBe(window.location.origin)
    // The key is kept for the next visit, and not printed in full.
    expect(localStorage.getItem('portal_key')).toBe('sk-0123456789abcdef')
    expect(wrapper.text()).not.toContain('sk-0123456789abcdef')
  })

  it('says so when the key is not known, and forgets it', async () => {
    localStorage.setItem('portal_key', 'sk-gone')
    fetchMock.mockImplementation(async () => json({ error: { message: 'Invalid API key' } }, 401))
    const wrapper = await mountApp()
    expect(wrapper.get('[data-test="portal-error"]').text()).toContain('not found')
    expect(wrapper.find('[data-test="portal-budget"]').exists()).toBe(false)
    expect(localStorage.getItem('portal_key')).toBeNull()
  })

  it('comes back with the saved key, tells a used-up budget apart, and signs out', async () => {
    localStorage.setItem('portal_key', 'sk-0123456789abcdef')
    fetchMock.mockImplementation(async (url: string) =>
      url.startsWith('/v1/models') ? json({}, 500) : json({ ...usage, status: 'quota_exhausted', expires_at: null, daily_usage: [], model_stats: [] })
    )
    const wrapper = await mountApp()
    expect(wrapper.get('[data-test="portal-status"]').text()).toBe('Budget used up')
    expect(wrapper.get('[data-test="portal-budget"]').text()).toContain('does not expire')
    expect(wrapper.find('[data-test="portal-daily-empty"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="portal-models"]').exists()).toBe(false)

    await wrapper.get('[data-test="portal-switch"]').trigger('click')
    expect(wrapper.find('[data-test="portal-enter"]').exists()).toBe(true)
    expect(localStorage.getItem('portal_key')).toBeNull()
  })

  it('switches language and remembers it', async () => {
    const wrapper = await mountApp()
    await wrapper.get('[data-test="portal-locale"]').trigger('click')
    expect(wrapper.text()).toContain('输入你的密钥')
    expect(localStorage.getItem('portal_locale')).toBe('zh')
  })
})
