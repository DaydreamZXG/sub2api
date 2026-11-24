<template>
  <AppLayout>
    <div class="ysb">
      <div v-if="loading" class="ysb-loading">
        <span class="ysb-spin"></span>
      </div>

      <div v-else-if="subscriptions.length === 0" data-test="subs-empty" class="ysb-empty">
        <span class="ysb-empty-icon"><Icon name="creditCard" size="lg" /></span>
        <h3>{{ t('userSubscriptions.noActiveSubscriptions') }}</h3>
        <p>{{ t(canBuy ? 'userSubscriptions.noActiveSubscriptionsBuy' : 'userSubscriptions.noActiveSubscriptionsDesc') }}</p>
        <router-link
          v-if="canBuy"
          :to="{ path: '/purchase', query: { tab: 'subscription' } }"
          class="ysb-btn ysb-btn-lime"
        >
          {{ t('userSubscriptions.browsePlans') }}
        </router-link>
      </div>

      <div v-else class="ysb-grid">
        <article
          v-for="subscription in subscriptions"
          :key="subscription.id"
          data-test="sub-card"
          class="ysb-card"
          :class="{ 'ysb-card-off': subscription.status !== 'active' }"
        >
          <header class="ysb-head">
            <div class="ysb-title">
              <h3>{{ subscription.group?.name || `Group #${subscription.group_id}` }}</h3>
              <span class="ysb-tag" :class="'ysb-tag-' + subscription.status">
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
            </div>
            <button
              v-if="subscription.status === 'active'"
              type="button"
              class="ysb-btn ysb-btn-lime"
              @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
            >
              {{ t('payment.renewNow') }}
            </button>
          </header>

          <p v-if="subscription.group?.description" class="ysb-desc">{{ subscription.group.description }}</p>

          <ul class="ysb-chips">
            <li v-if="subscription.group?.platform">{{ platformLabel(subscription.group.platform) }}</li>
            <li data-test="sub-pricing">{{ pricingLabel(subscription) }}</li>
            <li v-if="subscriptionHasPeakRate(subscription)" data-test="sub-peak" class="ysb-chip-peak">
              <Icon name="clock" size="xs" />{{ subscriptionPeakRateLabel(subscription) }}
            </li>
          </ul>

          <dl class="ysb-expiry">
            <dt>{{ t('userSubscriptions.expires') }}</dt>
            <dd v-if="subscription.expires_at" :class="getExpirationClass(subscription.expires_at)">
              {{ formatExpirationDate(subscription.expires_at) }}
            </dd>
            <dd v-else>{{ t('userSubscriptions.noExpiration') }}</dd>
          </dl>

          <div v-if="meters(subscription).length" class="ysb-meters">
            <div v-for="meter in meters(subscription)" :key="meter.key" data-test="sub-meter" class="ysb-meter">
              <div class="ysb-meter-top">
                <span class="ysb-meter-name">{{ t(meter.label) }}</span>
                <span class="ysb-meter-nums">
                  {{ sign }}{{ meter.used.toFixed(2) }} / {{ sign }}{{ meter.limit.toFixed(2) }}
                </span>
              </div>
              <div class="ysb-bar">
                <span :class="getProgressBarClass(meter.used, meter.limit)" :style="{ width: getProgressWidth(meter.used, meter.limit) }"></span>
              </div>
              <p v-if="meter.note" class="ysb-meter-note">{{ meter.note }}</p>
            </div>
          </div>

          <div v-else data-test="sub-unlimited" class="ysb-unlimited">
            <span class="ysb-unlimited-mark">∞</span>
            <div>
              <b>{{ t('userSubscriptions.unlimited') }}</b>
              <p>{{ t('userSubscriptions.unlimitedDesc') }}</p>
            </div>
          </div>
        </article>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTimeToMinute } from '@/utils/format'
import { hasPeakRate, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformLabel } from '@/utils/platformColors'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import {
  getExpirationDateRelation,
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts
} from '@/utils/subscriptionQuota'
import { ycSign as sign } from '@/composables/ycMoneyState'


const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

// "高峰时段 14:00-18:00 (UTC+8) 贵 20%": the same words the model plaza uses.
function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  const group = subscription.group
  if (!group) return ''
  const zone = serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  const hours = `${group.peak_start}-${group.peak_end}`
  return t('modelPlaza.detail.peakNote', {
    window: zone ? `${hours} (${zone})` : hours,
    change: priceChangeLabel(t, group.peak_rate_multiplier)
  })
}

// The plans are sold on the top-up page, which only exists while payment is on.
const canBuy = computed(() => isFeatureFlagEnabled(FeatureFlags.payment))

function pricingLabel(subscription: UserSubscription): string {
  const change = priceChangeLabel(t, subscription.group?.rate_multiplier ?? 1)
  return t('userSubscriptions.pricing', { change: change || t('userSubscriptions.standardPrice') })
}

interface Meter {
  key: string
  label: string
  used: number
  limit: number
  note: string
}

function meters(subscription: UserSubscription): Meter[] {
  const group = subscription.group
  const list: Meter[] = []
  if (group?.daily_limit_usd) {
    list.push({
      key: 'daily',
      label: 'userSubscriptions.daily',
      used: subscription.daily_usage_usd || 0,
      limit: group.daily_limit_usd,
      note: subscription.daily_window_start ? formatDailyUsageWindow(subscription) : ''
    })
  }
  if (group?.weekly_limit_usd) {
    list.push({
      key: 'weekly',
      label: 'userSubscriptions.weekly',
      used: subscription.weekly_usage_usd || 0,
      limit: group.weekly_limit_usd,
      note: subscription.weekly_window_start
        ? t('userSubscriptions.resetIn', { time: formatResetTime(subscription.weekly_window_start, 168) })
        : ''
    })
  }
  if (group?.monthly_limit_usd) {
    list.push({
      key: 'monthly',
      label: 'userSubscriptions.monthly',
      used: subscription.monthly_usage_usd || 0,
      limit: group.monthly_limit_usd,
      note: subscription.monthly_window_start
        ? t('userSubscriptions.resetIn', { time: formatResetTime(subscription.monthly_window_start, 720) })
        : ''
    })
  }
  return list
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return ''
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'ysb-bar-bad'
  if (percentage >= 70) return 'ysb-bar-warn'
  return ''
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const relation = getExpirationDateRelation(expires, now)

  if (relation === null) return ''

  if (relation === 'expired') {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (relation === 'today') {
    return `${dateStr} (${t('common.today')})`
  }
  if (relation === 'tomorrow') {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (diff <= 0) return 'ysb-bad'
  if (days <= 3) return 'ysb-bad'
  if (days <= 7) return 'ysb-warn'
  return ''
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return t('userSubscriptions.duration.daysHours', { d: parts.days, h: parts.hours })
  }

  if (parts.hours > 0) {
    return t('userSubscriptions.duration.hoursMinutes', { h: parts.hours, m: parts.minutes })
  }

  return t('userSubscriptions.duration.minutes', { m: parts.minutes })
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>

<style scoped>
.ysb { max-width: 1040px; }
.ysb-loading { display: grid; place-items: center; padding: 64px 0; }
.ysb-spin {
  width: 28px;
  height: 28px;
  border: 2px solid var(--yc-line);
  border-top-color: var(--yc-ink);
  border-radius: 50%;
  animation: ysb-spin 0.8s linear infinite;
}
@keyframes ysb-spin { to { transform: rotate(360deg); } }

.ysb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 650;
  white-space: nowrap;
}
.ysb-btn-lime { background: var(--yc-lime); color: var(--yc-on-lime); }
.ysb-btn-lime:hover { filter: brightness(0.96); }

.ysb-empty {
  display: grid;
  justify-items: center;
  gap: 10px;
  padding: 56px 24px;
  border: 1px dashed var(--yc-line);
  border-radius: 22px;
  background: var(--yc-paper);
  text-align: center;
}
.ysb-empty-icon { display: grid; place-items: center; width: 56px; height: 56px; border-radius: 18px; background: var(--yc-fill); color: var(--yc-mut); }
.ysb-empty h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.ysb-empty p { max-width: 420px; font-size: 14px; color: var(--yc-mut); }
.ysb-empty .ysb-btn { margin-top: 8px; }

.ysb-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 20px; align-items: start; }
.ysb-card {
  display: grid;
  gap: 16px;
  padding: 24px;
  border: 1px solid var(--yc-line);
  border-radius: 22px;
  background: var(--yc-paper);
}
.ysb-card-off { opacity: 0.72; }
.ysb-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.ysb-title { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; min-width: 0; }
.ysb-title h3 { min-width: 0; overflow-wrap: anywhere; font-size: 20px; font-weight: 750; letter-spacing: -0.01em; color: var(--yc-ink); }
.ysb-tag { padding: 2px 10px; border-radius: 999px; background: var(--yc-fill); color: var(--yc-mut); font-size: 12px; font-weight: 650; white-space: nowrap; }
.ysb-tag-active { background: var(--yc-lime); color: var(--yc-on-lime); }
.ysb-tag-revoked, .ysb-tag-suspended { background: color-mix(in srgb, #ef4444 16%, transparent); color: #dc2626; }
.ysb-desc { margin-top: -8px; font-size: 14px; color: var(--yc-mut); }

.ysb-chips { display: flex; flex-wrap: wrap; gap: 6px; margin: 0; padding: 0; list-style: none; }
.ysb-chips li {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border: 1px solid var(--yc-line);
  border-radius: 999px;
  color: var(--yc-mut);
  font-size: 12px;
}
.ysb-chips .ysb-chip-peak { border-color: color-mix(in srgb, #f59e0b 50%, transparent); color: #b45309; }
.dark .ysb-chips .ysb-chip-peak { color: #fbbf24; }

.ysb-expiry { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin: 0; padding: 12px 16px; border-radius: 14px; background: var(--yc-fill); font-size: 14px; }
.ysb-expiry dt { color: var(--yc-mut); white-space: nowrap; }
.ysb-expiry dd { margin: 0; color: var(--yc-ink); font-weight: 650; text-align: right; }
.ysb-expiry .ysb-warn { color: #d97706; }
.ysb-expiry .ysb-bad { color: #dc2626; }

.ysb-meters { display: grid; gap: 14px; }
.ysb-meter-top { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; font-size: 14px; }
.ysb-meter-name { font-weight: 650; color: var(--yc-ink); }
.ysb-meter-nums { color: var(--yc-mut); font-variant-numeric: tabular-nums; white-space: nowrap; }
.ysb-bar { height: 8px; margin-top: 8px; overflow: hidden; border-radius: 999px; background: var(--yc-fill); }
.ysb-bar span { display: block; height: 100%; border-radius: inherit; background: var(--yc-ink); transition: width 0.3s; }
.ysb-bar .ysb-bar-warn { background: #f59e0b; }
.ysb-bar .ysb-bar-bad { background: #ef4444; }
.ysb-meter-note { margin-top: 6px; font-size: 12px; color: var(--yc-mut); }

.ysb-unlimited { display: flex; align-items: center; gap: 14px; padding: 16px; border-radius: 14px; background: var(--yc-fill); }
.ysb-unlimited-mark { font-size: 34px; line-height: 1; color: var(--yc-ink); }
.ysb-unlimited b { font-size: 14px; color: var(--yc-ink); }
.ysb-unlimited p { font-size: 12px; color: var(--yc-mut); }

@media (max-width: 900px) {
  .ysb-grid { grid-template-columns: minmax(0, 1fr); }
  .ysb-card { padding: 20px; }
}
</style>
