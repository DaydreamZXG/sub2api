import { ref } from 'vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore, useAuthStore } from '@/stores'

// Balance units credited per ¥1 paid. Users see what they paid, so amounts are shown in ¥.
const unitsPerYuan = ref(1)
// Reference rate used only to restate official USD prices in ¥ for comparison.
const cnyPerUsd = ref(7.2)
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
