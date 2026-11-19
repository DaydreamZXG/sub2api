import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import YcGrowthCard from '../YcGrowthCard.vue'

const { getMyGrowth } = vi.hoisted(() => ({ getMyGrowth: vi.fn() }))

vi.mock('@/api/growth', () => ({ getMyGrowth }))
vi.mock('@/composables/useYcMoney', () => ({ useYcMoney: () => ({ yuan: (units: number) => `¥${units}` }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string, params?: object) => (params ? `${key} ${JSON.stringify(params)}` : key) })
}))

const levels = [
  { name: 'Seed', threshold: 0, discount: 0 },
  { name: 'Sprout', threshold: 100, discount: 0.02 },
  { name: 'Harvest', threshold: 500, discount: 0.125 }
]

const mountCard = async (growth: object) => {
  getMyGrowth.mockResolvedValue(growth)
  const wrapper = mount(YcGrowthCard)
  await flushPromises()
  return wrapper
}

describe('YcGrowthCard', () => {
  beforeEach(() => getMyGrowth.mockReset())

  it('stays hidden while growth levels are off', async () => {
    const wrapper = await mountCard({ enabled: false, levels: [], level_index: -1, total_recharged: 0 })

    expect(wrapper.find('.yc-growth').exists()).toBe(false)
  })

  it('shows the level, its discount and what is left to the next one', async () => {
    const wrapper = await mountCard({ enabled: true, levels, level_index: 1, total_recharged: 200 })

    expect(wrapper.find('.yc-growth-name').text()).toContain('Sprout')
    expect(wrapper.find('.yc-growth-name').text()).toContain('{"percent":"2"}')
    expect(wrapper.find('.yc-growth-next').text()).toBe('yc.growth.next {"amount":"¥300","name":"Harvest","percent":"12.5"}')
    expect(wrapper.find('[role="progressbar"]').attributes('aria-valuenow')).toBe('25')
  })

  it('has no next step at the top level', async () => {
    const wrapper = await mountCard({ enabled: true, levels, level_index: 2, total_recharged: 900 })

    expect(wrapper.find('.yc-growth-next').text()).toBe('yc.growth.top')
    expect(wrapper.find('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')
  })
})
