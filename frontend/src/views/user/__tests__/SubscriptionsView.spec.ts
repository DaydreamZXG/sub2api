import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SubscriptionsView from '@/views/user/SubscriptionsView.vue'

const { getMySubscriptions, settings } = vi.hoisted(() => ({
  getMySubscriptions: vi.fn(),
  settings: { payment_enabled: true as boolean }
}))

vi.mock('@/api/subscriptions', () => ({ default: { getMySubscriptions } }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: settings, showError: vi.fn() })
}))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ cachedPublicSettings: settings, showError: vi.fn() })
}))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, named?: Record<string, unknown>) => (named ? `${key} ${JSON.stringify(named)}` : key)
    })
  }
})

function subscription(group: Record<string, unknown>, extra: Record<string, unknown> = {}) {
  return {
    id: 1,
    group_id: 9,
    status: 'active',
    expires_at: null,
    daily_usage_usd: 0,
    weekly_usage_usd: 0,
    monthly_usage_usd: 0,
    group: { name: 'Claude 包月', platform: 'anthropic', rate_multiplier: 1, ...group },
    ...extra
  }
}

async function mountWith(list: unknown[]) {
  getMySubscriptions.mockResolvedValue(list)
  const wrapper = mount(SubscriptionsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        RouterLink: { template: '<a data-test="plans-link"><slot /></a>' },
        Icon: true
      }
    }
  })
  await flushPromises()
  return wrapper
}

describe('SubscriptionsView', () => {
  beforeEach(() => {
    getMySubscriptions.mockReset()
    settings.payment_enabled = true
  })

  it('says the price in words and never shows the raw multiplier', async () => {
    const wrapper = await mountWith([subscription({ rate_multiplier: 0.8 })])

    const pricing = wrapper.get('[data-test="sub-pricing"]').text()
    expect(pricing).toContain('yc.priceWords.cheaper')
    expect(wrapper.text()).not.toContain('×')
    expect(wrapper.text()).not.toContain('payment.planCard.rate')
  })

  it('calls an unchanged price the standard price', async () => {
    const wrapper = await mountWith([subscription({ rate_multiplier: 1 })])

    expect(wrapper.get('[data-test="sub-pricing"]').text()).toContain('userSubscriptions.standardPrice')
  })

  it('describes peak hours as a price change', async () => {
    const wrapper = await mountWith([
      subscription({ peak_rate_enabled: true, peak_start: '14:00', peak_end: '18:00', peak_rate_multiplier: 1.2 })
    ])

    const peak = wrapper.get('[data-test="sub-peak"]').text()
    expect(peak).toContain('14:00-18:00')
    expect(peak).toContain('yc.priceWords.dearer')
    expect(peak).not.toContain('×')
  })

  it('shows one meter per configured limit, and the unlimited note when there is none', async () => {
    const limited = await mountWith([
      subscription({ daily_limit_usd: 10, monthly_limit_usd: 100 }, { daily_usage_usd: 9.5, monthly_usage_usd: 20 })
    ])
    const meters = limited.findAll('[data-test="sub-meter"]')
    expect(meters).toHaveLength(2)
    expect(meters[0].text()).toContain('9.50')
    expect(meters[0].find('.ysb-bar-bad').exists()).toBe(true)
    expect(meters[1].find('.ysb-bar-bad, .ysb-bar-warn').exists()).toBe(false)
    expect(limited.find('[data-test="sub-unlimited"]').exists()).toBe(false)

    const unlimited = await mountWith([subscription({})])
    expect(unlimited.findAll('[data-test="sub-meter"]')).toHaveLength(0)
    expect(unlimited.find('[data-test="sub-unlimited"]').exists()).toBe(true)
  })

  it('offers the plans from the empty state only while payment is on', async () => {
    const on = await mountWith([])
    expect(on.get('[data-test="subs-empty"]').text()).toContain('userSubscriptions.noActiveSubscriptionsBuy')
    expect(on.find('[data-test="plans-link"]').exists()).toBe(true)

    settings.payment_enabled = false
    const off = await mountWith([])
    expect(off.get('[data-test="subs-empty"]').text()).toContain('userSubscriptions.noActiveSubscriptionsDesc')
    expect(off.find('[data-test="plans-link"]').exists()).toBe(false)
  })
})
