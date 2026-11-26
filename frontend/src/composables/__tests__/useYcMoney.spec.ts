import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick, reactive } from 'vue'

const appStore = reactive<{ cachedPublicSettings: Record<string, unknown> | null }>({ cachedPublicSettings: null })
const authStore = reactive({ isAuthenticated: false })
const getCheckoutInfo = vi.fn()

vi.mock('@/stores', () => ({ useAppStore: () => appStore, useAuthStore: () => authStore }))
vi.mock('@/api/payment', () => ({ paymentAPI: { getCheckoutInfo: () => getCheckoutInfo() } }))

async function setup() {
  vi.resetModules()
  const state = await import('../ycMoneyState')
  const { useYcMoney } = await import('../useYcMoney')
  const scope = effectScope()
  const money = scope.run(() => useYcMoney())!
  return { state, money, scope }
}

describe('useYcMoney', () => {
  beforeEach(() => {
    appStore.cachedPublicSettings = null
    authStore.isAuthenticated = false
    getCheckoutInfo.mockReset()
  })

  it('gives a visitor the site rate from the public settings, without asking for checkout info', async () => {
    appStore.cachedPublicSettings = {
      payment_enabled: true,
      balance_recharge_multiplier: 7,
      subscription_usd_to_cny_rate: 7.3,
      reference_usd_to_cny_rate: 7.1
    }
    const { state, money, scope } = await setup()

    expect(state.unitsPerDollar.value).toBe(7)
    expect(money.dollars(14)).toBe('$2.00')
    expect(getCheckoutInfo).not.toHaveBeenCalled()
    scope.stop()
  })

  it('picks the rates up when the public settings arrive after the page', async () => {
    const { state, scope } = await setup()
    expect(state.unitsPerDollar.value).toBe(1)

    appStore.cachedPublicSettings = { payment_enabled: true, balance_recharge_multiplier: 5, subscription_usd_to_cny_rate: 0 }
    await nextTick()

    expect(state.unitsPerDollar.value).toBe(5)
    scope.stop()
  })

  it('keeps the defaults when payments are off or the server sends no rates', async () => {
    appStore.cachedPublicSettings = { payment_enabled: false, balance_recharge_multiplier: 7, subscription_usd_to_cny_rate: 7.3 }
    const off = await setup()
    expect(off.state.unitsPerDollar.value).toBe(1)
    off.scope.stop()

    appStore.cachedPublicSettings = { payment_enabled: true }
    const older = await setup()
    expect(older.state.unitsPerDollar.value).toBe(1)
    older.scope.stop()
  })

})
