import { paymentAPI } from '@/api/payment'
import { useAppStore, useAuthStore } from '@/stores'
import { cnyPerUsd, unitsPerYuan } from './ycMoneyState'

let requested = false

export function useYcMoney() {
  const appStore = useAppStore()
  const authStore = useAuthStore()

  if (!requested && authStore.isAuthenticated && appStore.cachedPublicSettings?.payment_enabled) {
    requested = true
    paymentAPI
      .getCheckoutInfo()
      .then(({ data }) => {
        if (data.balance_recharge_multiplier > 0) unitsPerYuan.value = data.balance_recharge_multiplier
        if (data.subscription_usd_to_cny_rate > 0) cnyPerUsd.value = data.subscription_usd_to_cny_rate
      })
      .catch(() => {
        requested = false
      })
  }

  function toYuan(units: number | null | undefined): number {
    return (units ?? 0) / unitsPerYuan.value
  }
  function yuan(units: number | null | undefined, digits = 2): string {
    return `¥${toYuan(units).toFixed(digits)}`
  }

  return { toYuan, yuan, cnyPerUsd }
}
