<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.topUp.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.topUp.lead') }}</p>

    <section class="yc-surface yc-top">
      <div>
        <p class="yc-top-lab">{{ t('yc.dash.balance') }}</p>
        <p class="yc-top-num yc-num"><span class="yc-mark">{{ yuan(authStore.user?.balance) }}</span></p>
      </div>
      <div class="yc-top-cta">
        <router-link v-if="canPay" to="/purchase" class="yc-btn yc-btn-lime">{{ t('yc.topUp.pay') }} →</router-link>
        <router-link v-else-if="payEnabled" to="/purchase" class="yc-btn yc-btn-lime">{{ t('yc.topUp.buyPlan') }} →</router-link>
        <router-link to="/redeem" class="yc-btn yc-btn-ghost">{{ t('yc.topUp.redeem') }}</router-link>
      </div>
    </section>

    <section class="yc-top-facts">
      <div v-for="f in facts" :key="f.title" class="yc-surface yc-top-fact">
        <h2>{{ f.title }}</h2>
        <p class="yc-note">{{ f.text }}</p>
      </div>
    </section>

    <p class="yc-note yc-top-foot">
      <router-link v-if="payEnabled" to="/orders" class="yc-link">{{ t('yc.topUp.orders') }} →</router-link>
    </p>
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import type { CheckoutInfoResponse } from '@/types/payment'
import YcShell from '@/components/yecai/YcShell.vue'
import { useYcMoney } from '@/composables/useYcMoney'
import { currencySymbol, normalizePaymentCurrency } from '@/components/payment/currency'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { yuan } = useYcMoney()

const checkout = ref<CheckoutInfoResponse | null>(null)
const payEnabled = computed(() => appStore.cachedPublicSettings?.payment_enabled === true)
const canPay = computed(() => payEnabled.value && checkout.value?.balance_disabled !== true)

// Limits and bonus thresholds are counted in the payment method's currency. Amounts are only
// quoted when every method on offer uses the same one.
const payCurrency = computed(() => {
  const used = new Set(
    Object.values(checkout.value?.methods ?? {}).map((method) => normalizePaymentCurrency(method.currency))
  )
  return used.size === 1 ? [...used][0] : null
})

const facts = computed(() => {
  const info = checkout.value
  const currency = payCurrency.value
  const paid = (amount: number) => `${currencySymbol(currency)}${amount}`
  const tiers = currency && info?.recharge_bonus_mode !== 'discount' ? (info?.recharge_bonus_tiers ?? []) : []
  return [
    {
      title: t('yc.topUp.minTitle'),
      text: info && currency && info.global_min > 0 ? t('yc.topUp.minText', { min: paid(info.global_min) }) : t('yc.topUp.minAny')
    },
    ...tiers.map((tier) => ({
      title: t('yc.topUp.bonusTitle', { pct: tier.bonus_percent }),
      text: t('yc.topUp.bonusText', { min: paid(tier.min_amount), pct: tier.bonus_percent })
    })),
    { title: t('yc.topUp.refundTitle'), text: t('yc.topUp.refundText') },
    { title: t('yc.topUp.clearTitle'), text: t('yc.topUp.clearText') }
  ]
})

onMounted(async () => {
  authStore.refreshUser().catch(() => undefined)
  if (!payEnabled.value) return
  try {
    checkout.value = (await paymentAPI.getCheckoutInfo()).data
  } catch (error) {
    console.error('Failed to load top-up info:', error)
  }
})
</script>

<style scoped>
.yc-top { display: flex; flex-wrap: wrap; align-items: flex-end; justify-content: space-between; gap: 20px; margin-top: 24px; padding: 28px 30px; }
.yc-top-lab { font-size: 13px; font-weight: 650; color: var(--yc-mut); }
.yc-top-num { margin-top: 6px; font-size: clamp(44px, 6vw, 72px); line-height: 1.1; }
.yc-top-cta { display: flex; flex-wrap: wrap; gap: 10px; }
.yc-top-facts { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 16px; margin-top: 16px; }
.yc-top-fact { padding: 22px 24px; }
.yc-top-fact h2 { font-size: 17px; font-weight: 750; letter-spacing: -0.01em; }
.yc-top-fact .yc-note { margin-top: 6px; }
.yc-top-foot { margin-top: 16px; }
@media (max-width: 760px) {
  .yc-top { padding: 22px 20px; }
}
</style>
