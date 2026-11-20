import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PlazaFilterBar from '../PlazaFilterBar.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const groups = [
  { id: 1, name: 'Claude line', platform: 'anthropic' },
  { id: 2, name: 'GPT line', platform: 'openai' },
]

function mountBar(props: Partial<{ platform: string; groupId: number | 'all' }> = {}) {
  return mount(PlazaFilterBar, {
    props: { platforms: ['anthropic', 'openai'], groups, platform: 'all', groupId: 'all', search: '', ...props },
    global: { stubs: { Icon: true, PlatformIcon: true } },
  })
}

describe('PlazaFilterBar', () => {
  it('filters by platform, line and name only, with no multiplier row', () => {
    const wrapper = mountBar()
    expect(wrapper.findAll('.pzf-row')).toHaveLength(2)
    expect(wrapper.text()).not.toContain('modelPlaza.filters.rateLabel')
    expect(wrapper.find('input[type="search"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('greys out lines of another platform and emits the chosen line', async () => {
    const wrapper = mountBar({ platform: 'openai' })
    const chips = wrapper.findAll('.pzf-row')[1].findAll('button')
    expect(chips.map((chip) => chip.attributes('disabled') !== undefined)).toEqual([false, true, false])
    await chips[2].trigger('click')
    expect(wrapper.emitted('update:groupId')).toEqual([[2]])
    wrapper.unmount()
  })
})
