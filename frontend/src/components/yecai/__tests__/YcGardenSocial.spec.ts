import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import YcGardenSocial from '../YcGardenSocial.vue'

const api = vi.hoisted(() => ({
  feedDog: vi.fn(),
  getGardenFeed: vi.fn(),
  getGardenNeighbours: vi.fn(),
  getNeighbourGarden: vi.fn(),
  stealCrop: vi.fn(),
  waterCrop: vi.fn()
}))
const toast = vi.hoisted(() => ({ showSuccess: vi.fn(), showError: vi.fn() }))

vi.mock('@/api/garden', () => api)
vi.mock('@/stores', () => ({ useAppStore: () => toast }))
vi.mock('@/composables/useYcMoney', () => ({ useYcMoney: () => ({ yuan: (units: number) => `¥${units}` }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string, params?: object) => (params ? `${key} ${JSON.stringify(params)}` : key),
    te: (key: string) => key === 'yc.garden.errors.GARDEN_NEED_CROP',
    locale: { value: 'en' }
  })
}))

const NOW = Date.parse('2026-11-20T12:00:00Z')
const at = (hours: number) => new Date(NOW + hours * 3600_000).toISOString()
const crop = (id: number, over: object) => ({
  id, size: 'small', planted_at: at(-30), ripe_at: at(-6), open_at: at(-4), can_water: false, can_steal: false, tried: '', ...over
})
const garden = (over: object = {}) => ({
  enabled: true, ripen_hours: 24, protect_hours: 2, steal_percent: 3, max_stolen_percent: 50, daily_steal_limit: 10,
  water_minutes: 60, max_waters_per_crop: 5, neighbours_per_day: 10, dog_min_percent: 10, dog_max_percent: 50,
  now: at(0), crops: [], dog_fed_today: false, dog_catch_percent: 10, steals_today: 0, unread: 0, ...over
})

const visit = async (crops: object[], over: object = {}) => {
  api.getGardenNeighbours.mockResolvedValue([{ id: 2, name: 'Bob', pinned: true, crops: crops.length }])
  api.getNeighbourGarden.mockResolvedValue({ neighbour: { id: 2, name: 'Bob', pinned: true, crops: crops.length }, crops, now: at(0) })
  api.getGardenFeed.mockResolvedValue([])
  const wrapper = mount(YcGardenSocial, { props: { garden: garden(over), now: NOW }, global: { stubs: { YcPlant: true } } })
  await flushPromises()
  await wrapper.find('.yc-nb-row').trigger('click')
  await flushPromises()
  return wrapper
}

describe('YcGardenSocial', () => {
  beforeEach(() => {
    Object.values(api).forEach((fn) => fn.mockReset())
    Object.values(toast).forEach((fn) => fn.mockReset())
  })

  it('offers watering on a growing crop and taking on an open one, and nothing on the rest', async () => {
    const wrapper = await visit([
      crop(1, { ripe_at: at(5), open_at: at(7), can_water: true }),
      crop(2, { can_steal: true }),
      crop(3, { ripe_at: at(-1), open_at: at(1) }),
      crop(4, { tried: 'caught' })
    ])

    const crops = wrapper.findAll('.yc-nb-crop')
    expect(crops.map((c) => c.find('button').exists() && c.find('button').text())).toEqual(['yc.garden.water', 'yc.garden.steal', false, false])
    expect(crops.map((c) => c.find('.yc-nb-state').text())).toEqual([
      'yc.garden.growing', 'yc.garden.state.open', 'yc.garden.state.protected', 'yc.garden.state.caught'
    ])
  })

  it('reports what was taken, and tells the parent to refresh', async () => {
    api.stealCrop.mockResolvedValue({ caught: false, amount: 6 })
    const wrapper = await visit([crop(2, { can_steal: true })])

    await wrapper.find('.yc-nb-crop button').trigger('click')
    await flushPromises()

    expect(api.stealCrop).toHaveBeenCalledWith(2)
    expect(toast.showSuccess).toHaveBeenCalledWith('yc.garden.stealDone {"amount":"¥6"}')
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })

  it('says so when the dog catches the thief', async () => {
    api.stealCrop.mockResolvedValue({ caught: true, amount: 0 })
    const wrapper = await visit([crop(2, { can_steal: true })])

    await wrapper.find('.yc-nb-crop button').trigger('click')
    await flushPromises()

    expect(toast.showError).toHaveBeenCalledWith('yc.garden.stealCaught')
    expect(toast.showSuccess).not.toHaveBeenCalled()
  })

  it('explains a refusal by its reason, and disables taking once the daily tries are used up', async () => {
    api.stealCrop.mockRejectedValue({ reason: 'GARDEN_NEED_CROP' })
    const wrapper = await visit([crop(2, { can_steal: true })])
    await wrapper.find('.yc-nb-crop button').trigger('click')
    await flushPromises()
    expect(toast.showError).toHaveBeenCalledWith('yc.garden.errors.GARDEN_NEED_CROP')

    const spent = await visit([crop(2, { can_steal: true })], { steals_today: 10 })
    expect(spent.find('.yc-nb-crop button').attributes('disabled')).toBeDefined()
  })

  it('offers to feed a hungry dog only', async () => {
    const hungry = await visit([])
    expect(hungry.find('.yc-dog button').text()).toBe('yc.garden.dog.feed')
    await hungry.find('.yc-dog button').trigger('click')
    await flushPromises()
    expect(api.feedDog).toHaveBeenCalledOnce()

    const fed = await visit([], { dog_fed_today: true, dog_catch_percent: 30 })
    expect(fed.find('.yc-dog button').exists()).toBe(false)
    expect(fed.find('.yc-dog-text p').text()).toBe('yc.garden.dog.fed {"percent":30}')
  })
})
