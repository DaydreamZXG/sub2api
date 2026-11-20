import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import YcAgentGuideView from '../YcAgentGuideView.vue'

const settings = vi.hoisted(() => ({ value: {} as { contact_info?: string } }))

vi.mock('@/stores', () => ({ useAppStore: () => ({ cachedPublicSettings: settings.value }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

const stubs = {
  YcShell: { template: '<div><slot /></div>' },
  RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }
}
const mountView = () => mount(YcAgentGuideView, { global: { stubs } })

describe('YcAgentGuideView', () => {
  beforeEach(() => {
    settings.value = {}
  })

  it('explains the terms, how customers come under an agent, the rules and the payout', () => {
    const wrapper = mountView()
    expect(wrapper.findAll('[data-test="agent-guide-terms"] dt')).toHaveLength(4)
    expect(wrapper.get('[data-test="agent-guide-terms"]').text()).toContain('yc.agentGuide.example')
    expect(wrapper.findAll('[data-test="agent-guide-join"] li')).toHaveLength(2)
    expect(wrapper.findAll('[data-test="agent-guide-rules"] li')).toHaveLength(6)
    expect(wrapper.findAll('[data-test="agent-guide-payout"] li')).toHaveLength(3)
  })

  it('shows how to reach the site, or sends the reader to Help when no contact is set', () => {
    const apply = mountView().get('[data-test="agent-guide-apply"]')
    expect(apply.find('.yc-help-contact').exists()).toBe(false)
    expect(apply.get('a').attributes('href')).toBe('/help')

    settings.value = { contact_info: 'WeChat: yecai' }
    const withContact = mountView().get('[data-test="agent-guide-apply"]')
    expect(withContact.get('.yc-help-contact').text()).toBe('WeChat: yecai')
    expect(withContact.find('a').exists()).toBe(false)
  })
})
