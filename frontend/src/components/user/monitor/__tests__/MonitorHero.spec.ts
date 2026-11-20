import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import MonitorHero from '../MonitorHero.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

function mountHero(overallStatus: 'operational' | 'degraded') {
  return mount(MonitorHero, {
    props: { overallStatus, intervalSeconds: 60, window: '7d', loading: false },
    global: { stubs: { Icon: true, AutoRefreshButton: true } },
  })
}

describe('MonitorHero', () => {
  it('says the overall state as a sentence instead of a status code', () => {
    const ok = mountHero('operational')
    expect(ok.get('[data-test="mon-headline"]').text()).toBe('channelStatus.headline.operational')
    expect(ok.text()).toContain('channelStatus.headlineHint.operational')
    expect(ok.text()).not.toContain('channelStatus.overall')

    const bad = mountHero('degraded')
    expect(bad.get('[data-test="mon-headline"]').text()).toBe('channelStatus.headline.degraded')
    expect(bad.classes()).toContain('ych-hero-degraded')
  })

  it('marks the chosen window and reports a change or a refresh', async () => {
    const wrapper = mountHero('operational')
    const tabs = wrapper.findAll('[role="tab"]')
    expect(tabs).toHaveLength(3)
    expect(tabs[0].attributes('aria-selected')).toBe('true')

    await tabs[2].trigger('click')
    expect(wrapper.emitted('update:window')).toEqual([['30d']])

    await wrapper.get('.ych-refresh').trigger('click')
    expect(wrapper.emitted('refresh')).toHaveLength(1)
  })
})
