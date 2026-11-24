import { watch } from 'vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore, useAuthStore } from '@/stores'
import { cnyPerUsd, unitsPerYuan } from './ycMoneyState'

let requested = false

export function useYcMoney() {
  const appStore = useAppStore()
  const authStore = useAuthStore()

  // The public settings carry both rates, so visitors get the same yuan prices as signed-in users.
  function applyPublicRates() {
    const settings = appStore.cachedPublicSettings
    if ((settings?.reference_usd_to_cny_rate ?? 0) > 0) cnyPerUsd.value = settings!.reference_usd_to_cny_rate!
    if (!settings?.payment_enabled) return
    if ((settings.balance_recharge_multiplier ?? 0) > 0) unitsPerYuan.value = settings.balance_recharge_multiplier!
  }

  function loadRate() {
    applyPublicRates()
    if (requested || !authStore.isAuthenticated || !appStore.cachedPublicSettings?.payment_enabled) return
    requested = true
    paymentAPI
      .getCheckoutInfo()
      .then(({ data }) => {
        if (data.balance_recharge_multiplier > 0) unitsPerYuan.value = data.balance_recharge_multiplier
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
      appStore.cachedPublicSettings?.balance_recharge_multiplier,
      appStore.cachedPublicSettings?.reference_usd_to_cny_rate
    ],
    loadRate
  )

  function toYuan(units: number | null | undefined): number {
    return (units ?? 0) / unitsPerYuan.value
  }
  function yuan(units: number | null | undefined, digits = 2): string {
    return `¥${toYuan(units).toFixed(digits)}`
  }

  /** The other way round, for amounts an admin types in yuan. */
  function toUnits(yuanAmount: number | null | undefined): number {
    return (yuanAmount ?? 0) * unitsPerYuan.value
  }

  return { toYuan, toUnits, yuan, cnyPerUsd }
}
