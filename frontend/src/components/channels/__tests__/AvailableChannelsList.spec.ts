import { createPinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AvailableChannelsList from '../AvailableChannelsList.vue'
import type { UserAvailableChannel } from '@/api/channels'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        params && Object.keys(params).length > 0 ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

const rows: UserAvailableChannel[] = [
  {
    name: 'Primary channel',
    description: 'Fast and reliable access',
    platforms: [
      {
        platform: 'anthropic',
        groups: [
          {
            id: 2,
            name: 'Public',
            platform: 'anthropic',
            subscription_type: 'standard',
            rate_multiplier: 1,
            peak_rate_enabled: false,
            peak_start: '',
            peak_end: '',
            peak_rate_multiplier: 1,
            is_exclusive: false,
          },
          {
            id: 1,
            name: 'Exclusive Pro',
            platform: 'anthropic',
            subscription_type: 'subscription',
            rate_multiplier: 1.2,
            peak_rate_enabled: true,
            peak_start: '08:00',
            peak_end: '10:00',
            peak_rate_multiplier: 1.5,
            is_exclusive: true,
          },
        ],
        supported_models: [{ name: 'claude-test', platform: 'anthropic', pricing: null }],
      },
    ],
  },
]

const baseProps = {
  rows,
  loading: false,
  pricingKeyPrefix: 'availableChannels.pricing',
  noPricingLabel: 'No pricing',
  noModelsLabel: 'No models',
  emptyLabel: 'No channels',
  userGroupRates: { 1: 0.8 },
}

function mountList(props = {}) {
  return mount(AvailableChannelsList, {
    props: { ...baseProps, ...props },
    global: {
      plugins: [createPinia()],
      stubs: {
        Icon: { props: ['name'], template: '<i :data-icon="name" />' },
        PlatformIcon: { template: '<i data-platform-icon />' },
        SupportedModelChip: {
          props: ['model', 'noPricingLabel'],
          template: '<span data-model-chip>{{ model.name }}:{{ noPricingLabel }}</span>',
        },
      },
    },
  })
}

describe('AvailableChannelsList', () => {
  it('shows one card per channel with its platform, lines and models', () => {
    const wrapper = mountList()

    const cards = wrapper.findAll('[data-test="channel-card"]')
    expect(cards).toHaveLength(1)
    expect(cards[0].text()).toContain('Primary channel')
    expect(cards[0].text()).toContain('Fast and reliable access')
    expect(wrapper.findAll('[data-test="channel-platform"]')).toHaveLength(1)
    expect(wrapper.get('[data-model-chip]').text()).toBe('claude-test:No pricing')
  })

  it('lists the lines granted to the user before the public ones', () => {
    const lines = mountList().findAll('[data-test="channel-line"]')

    expect(lines).toHaveLength(2)
    expect(lines[0].text()).toContain('Exclusive Pro')
    expect(lines[0].text()).toContain('availableChannels.exclusive')
    expect(lines[0].text()).toContain('availableChannels.subscription')
    expect(lines[1].text()).toContain('Public')
    expect(lines[1].text()).toContain('availableChannels.public')
    expect(lines[1].text()).not.toContain('availableChannels.subscription')
  })

  it("says each line's price in words, using the user's own rate when there is one", () => {
    const prices = mountList().findAll('[data-test="channel-line-price"]')

    // Line 1 is 1.2 by default, but this user has 0.8 on it.
    expect(prices[0].text()).toBe('yc.priceWords.cheaper:{"zhe":8,"pct":20}')
    expect(prices[0].attributes('title')).toBe('modelPlaza.table.rawRate:{"rate":0.8}')
    expect(prices[1].text()).toBe('userSubscriptions.standardPrice')
    // Only a price below the standard one is highlighted.
    expect(prices[0].classes()).toContain('yav-price-good')
    expect(prices[1].classes()).not.toContain('yav-price-good')
    expect(mountList().text()).not.toContain('×')
  })

  it('falls back to the default rate of the line and marks a free line', () => {
    const wrapper = mountList({ userGroupRates: { 2: 0 } })
    const prices = wrapper.findAll('[data-test="channel-line-price"]')

    expect(prices[0].text()).toBe('yc.priceWords.dearer:{"pct":20}')
    expect(prices[1].text()).toBe('yc.priceWords.free')
  })

  it('writes the peak hours as a sentence only on lines that have them', () => {
    const wrapper = mountList()
    const peaks = wrapper.findAll('[data-test="channel-line-peak"]')

    expect(peaks).toHaveLength(1)
    expect(peaks[0].text()).toContain('modelPlaza.detail.peakNote')
    expect(peaks[0].text()).toContain('08:00-10:00')
    expect(peaks[0].text()).toContain('yc.priceWords.dearer')
    expect(peaks[0].attributes('title')).toContain('modelPlaza.table.rawRate:{"rate":1.5}')
  })

  it('keeps the placeholders when a platform has no lines or models', () => {
    const wrapper = mountList({
      rows: [
        {
          name: 'Fallback channel',
          description: '',
          platforms: [{ platform: 'openai', groups: [], supported_models: [] }],
        },
      ],
    })

    expect(wrapper.text()).toContain('Fallback channel')
    expect(wrapper.get('[data-test="channel-no-lines"]').text()).toBe('-')
    expect(wrapper.text()).toContain('No models')
    expect(wrapper.find('[data-test="channel-line"]').exists()).toBe(false)
    expect(wrapper.find('.yav-head p').exists()).toBe(false)
  })

  it('shows a spinner while loading and a message when there is nothing', async () => {
    const wrapper = mountList({ loading: true, rows: [] })

    expect(wrapper.get('[data-test="channels-loading"] [data-icon="refresh"]')).toBeTruthy()
    expect(wrapper.find('[data-test="channel-card"]').exists()).toBe(false)

    await wrapper.setProps({ loading: false })

    expect(wrapper.find('[data-test="channels-loading"]').exists()).toBe(false)
    expect(wrapper.get('[data-test="channels-empty"]').text()).toContain('No channels')
  })
})
