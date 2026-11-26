import { watch } from 'vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore, useAuthStore } from '@/stores'
import { unitsPerDollar } from './ycMoneyState'

let requested = false

export function useYcMoney() {
  const appStore = useAppStore()
  const authStore = useAuthStore()

  function applyPublicRate() {
    const settings = appStore.cachedPublicSettings
    if (!settings?.payment_enabled) return
    if ((settings.balance_recharge_multiplier ?? 0) > 0) unitsPerDollar.value = settings.balance_recharge_multiplier!
  }

  function loadRate() {
    applyPublicRate()
    if (requested || !authStore.isAuthenticated || !appStore.cachedPublicSettings?.payment_enabled) return
    requested = true
    paymentAPI
      .getCheckoutInfo()
      .then(({ data }) => {
        if (data.balance_recharge_multiplier > 0) unitsPerDollar.value = data.balance_recharge_multiplier
      })
      .catch(() => {
        requested = false
      })
  }
  loadRate()
  // Sign-in state and public settings can arrive after the first page has rendered.
  watch(
    () => [
      authStore.isAuthenticated,
      appStore.cachedPublicSettings?.payment_enabled,
      appStore.cachedPublicSettings?.balance_recharge_multiplier
    ],
    loadRate
  )

  function toDollars(units: number | null | undefined): number {
    return (units ?? 0) / unitsPerDollar.value
  }
  function dollars(units: number | null | undefined, digits = 2): string {
    return `$${toDollars(units).toFixed(digits)}`
  }

  /** The other way round, for amounts an admin types in dollars. */
  function toUnits(amount: number | null | undefined): number {
    return (amount ?? 0) * unitsPerDollar.value
  }

  return { toDollars, toUnits, dollars }
}
