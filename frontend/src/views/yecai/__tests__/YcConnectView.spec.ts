import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { connectDeepLink } from '@/api/desktop'
import YcConnectView from '../YcConnectView.vue'

const { issueConnectCode, stationOrigin, openDesktopApp, showError, showInfo } = vi.hoisted(() => ({
  issueConnectCode: vi.fn(),
  stationOrigin: vi.fn(),
  openDesktopApp: vi.fn(),
  showError: vi.fn(),
  showInfo: vi.fn()
}))

vi.mock('@/api/desktop', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/api/desktop')>()),
  issueConnectCode,
  stationOrigin,
  openDesktopApp
}))
vi.mock('@/api/keys', () => ({ keysAPI: { list: vi.fn().mockResolvedValue({ items: [] }) } }))
vi.mock('@/api/groups', () => ({ userGroupsAPI: { getAvailable: vi.fn().mockResolvedValue([]) } }))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ cachedPublicSettings: null, showError, showInfo, showSuccess: vi.fn() })
}))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

const clickConnect = async () => {
  const wrapper = mount(YcConnectView, {
    global: { stubs: { YcShell: { template: '<div><slot /></div>' }, UseKeyModal: true } }
  })
  await flushPromises()
  const button = wrapper.findAll('button').find((item) => item.text() === 'yc.connect.appConnect')
  await button!.trigger('click')
  await flushPromises()
}

describe('YcConnectView connect to app', () => {
  beforeEach(() => {
    for (const fn of [issueConnectCode, stationOrigin, openDesktopApp, showError, showInfo]) fn.mockReset()
    stationOrigin.mockReturnValue('https://yeschoy.test')
  })

  it('issues a code and opens the app with it', async () => {
    const connect = { code: 'c'.repeat(43), origin: 'https://yeschoy.test', expiresAt: 0, ttlSeconds: 300 }
    issueConnectCode.mockResolvedValue(connect)

    await clickConnect()

    expect(issueConnectCode).toHaveBeenCalledOnce()
    expect(openDesktopApp).toHaveBeenCalledWith(connect)
    expect(showInfo).toHaveBeenCalledWith('yc.connect.appConnectHint')
  })

  it('says so and does not open the app when no code could be issued', async () => {
    issueConnectCode.mockRejectedValue(new Error('offline'))

    await clickConnect()

    expect(openDesktopApp).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('yc.connect.appConnectFailed')
  })

  it('does not ask for a code when the station is not served over https', async () => {
    stationOrigin.mockReturnValue('http://localhost:8080')

    await clickConnect()

    expect(issueConnectCode).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('yc.connect.appConnectNeedsHttps')
  })

  it('builds the deep link the desktop app expects', () => {
    expect(connectDeepLink({ code: 'abc-DEF_123', origin: 'https://yeschoy.test' })).toBe(
      'yeschoy://connect?origin=https%3A%2F%2Fyeschoy.test&code=abc-DEF_123'
    )
  })
})
