<template>
  <YcShell>
    <h1 class="yc-dash-h1">{{ t(greetingKey) }}{{ name ? t('yc.dash.comma') + name : '' }}</h1>

    <router-link v-if="next" :to="next.to" class="yc-next">
      <span class="yc-next-tag">{{ t('yc.dash.nextTag') }}</span>
      <span class="yc-next-text">{{ t(next.text) }}</span>
      <span class="yc-next-go">{{ t(next.action) }} →</span>
    </router-link>

    <section class="yc-tiles">
      <div class="yc-tile yc-wallet">
        <p class="yc-tile-lab">{{ t('yc.dash.balance') }}</p>
        <p class="yc-tile-num yc-num">{{ yuan(shown.balance) }}</p>
        <p class="yc-tile-sub">{{ t('yc.dash.balanceHint') }}</p>
        <router-link :to="topUpPath" class="yc-btn yc-btn-sm yc-wallet-btn">{{ t('yc.shell.topUp') }} →</router-link>
      </div>
      <div class="yc-surface yc-tile">
        <p class="yc-tile-lab">{{ t('yc.dash.today') }}</p>
        <p class="yc-tile-num yc-num">{{ yuan(shown.today) }}</p>
        <p class="yc-tile-sub">{{ t('yc.dash.requests', { n: stats?.today_requests ?? 0 }) }}</p>
      </div>
      <div class="yc-surface yc-tile">
        <p class="yc-tile-lab">{{ t('yc.dash.total') }}</p>
        <p class="yc-tile-num yc-num">{{ yuan(shown.total) }}</p>
        <p class="yc-tile-sub">{{ t('yc.dash.requests', { n: stats?.total_requests ?? 0 }) }}</p>
      </div>
    </section>

    <section class="yc-surface yc-recent">
      <header>
        <h2>{{ t('yc.dash.recent') }}</h2>
        <router-link to="/bills">{{ t('yc.dash.allBills') }} →</router-link>
      </header>
      <p v-if="!loading && recent.length === 0" class="yc-empty">{{ t('yc.dash.empty') }}</p>
      <template v-for="log in recent" :key="log.id">
        <button type="button" class="yc-row" :aria-expanded="open === log.id" @click="open = open === log.id ? null : log.id">
          <span class="yc-row-time">{{ formatRelativeTime(log.created_at) }}</span>
          <span class="yc-row-model">{{ log.model }}</span>
          <b>{{ yuan(log.actual_cost, 4) }}</b>
          <i aria-hidden="true">›</i>
        </button>
        <dl v-if="open === log.id" class="yc-detail">
          <div v-for="d in details(log)" :key="d.label"><dt>{{ t(d.label) }}</dt><dd>{{ d.value }}</dd></div>
        </dl>
      </template>
      <p v-if="recent.length" class="yc-tip">{{ t('yc.dash.expandTip') }}</p>
    </section>
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useTransition } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores'
import { usageAPI, type UserDashboardStats } from '@/api/usage'
import type { UsageLog } from '@/types'
import { formatDateLocalInput, formatRelativeTime } from '@/utils/format'
import YcShell from '@/components/yecai/YcShell.vue'
import { useYcMoney } from '@/composables/useYcMoney'

const { t } = useI18n()
const authStore = useAuthStore()
const { yuan } = useYcMoney()

const stats = ref<UserDashboardStats | null>(null)
const recent = ref<UsageLog[]>([])
const loading = ref(true)
const open = ref<number | null>(null)

const balance = computed(() => authStore.user?.balance ?? 0)
const name = computed(() => authStore.user?.username || '')

// Headline amounts count up to their value when the data arrives.
const still = window.matchMedia('(prefers-reduced-motion: reduce)').matches
const ease = { duration: still ? 0 : 900, transition: [0.2, 0.7, 0.2, 1] as [number, number, number, number] }
const shown = reactive({
  balance: useTransition(balance, ease),
  today: useTransition(() => stats.value?.today_actual_cost ?? 0, ease),
  total: useTransition(() => stats.value?.total_actual_cost ?? 0, ease)
})
const topUpPath = '/topup'

const greetingKey = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return 'yc.dash.night'
  if (h < 12) return 'yc.dash.morning'
  if (h < 18) return 'yc.dash.afternoon'
  return 'yc.dash.evening'
})

// One next step at a time, gone once the account has made a request.
const next = computed(() => {
  if (!stats.value) return null
  if (balance.value <= 0 && stats.value.total_requests === 0) {
    return { to: topUpPath, text: 'yc.dash.nextTopUp', action: 'yc.shell.topUp' }
  }
  if (stats.value.total_api_keys === 0) return { to: '/keys', text: 'yc.dash.nextKey', action: 'yc.dash.nextKeyGo' }
  if (stats.value.total_requests === 0) return { to: '/keys', text: 'yc.dash.nextConnect', action: 'yc.dash.nextConnectGo' }
  return null
})

function count(value: number | null | undefined): string {
  return (value ?? 0).toLocaleString()
}
function seconds(ms: number | null): string {
  return ms == null ? '-' : `${(ms / 1000).toFixed(1)}s`
}
function details(log: UsageLog) {
  return [
    { label: 'yc.dash.dModel', value: log.model },
    { label: 'yc.dash.dInput', value: count(log.input_tokens) },
    { label: 'yc.dash.dOutput', value: count(log.output_tokens) },
    { label: 'yc.dash.dCacheWrite', value: count(log.cache_creation_tokens) },
    { label: 'yc.dash.dCacheRead', value: count(log.cache_read_tokens) },
    { label: 'yc.dash.dTiming', value: `${seconds(log.first_token_ms)} / ${seconds(log.duration_ms)}` },
    { label: 'yc.dash.dListCost', value: yuan(log.total_cost, 4) },
    { label: 'yc.dash.dMultiplier', value: `×${log.rate_multiplier}` }
  ]
}

onMounted(async () => {
  const today = formatDateLocalInput(new Date())
  const weekAgo = formatDateLocalInput(new Date(Date.now() - 6 * 86400000))
  try {
    await authStore.refreshUser()
    const [s, logs] = await Promise.all([usageAPI.getDashboardStats(), usageAPI.getByDateRange(weekAgo, today)])
    stats.value = s
    recent.value = logs.items.slice(0, 6)
  } catch (error) {
    console.error('Failed to load the home page:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.yc-dash-h1 { font-size: clamp(30px, 4.6vw, 48px); line-height: 1.12; letter-spacing: -0.035em; font-weight: 800; }
.yc-next { display: flex; align-items: center; gap: 14px; margin-top: 24px; padding: 14px 18px; border-radius: 16px; background: var(--yc-lime); color: var(--yc-on-lime); font-size: 15.5px; }
.yc-next-tag { flex: none; padding: 3px 10px; border-radius: 999px; background: var(--yc-on-lime); color: var(--yc-lime); font-size: 12px; font-weight: 700; }
.yc-next-text { flex: 1; font-weight: 600; }
.yc-next-go { flex: none; font-weight: 750; }
.yc-tiles { display: grid; grid-template-columns: 1.3fr 1fr 1fr; gap: 16px; margin-top: 28px; }
.yc-tile { padding: 26px 28px 28px; }
.yc-tile-lab { font-size: 13px; font-weight: 650; color: var(--yc-mut); }
.yc-tile-num { margin-top: 8px; font-size: clamp(36px, 4.2vw, 52px); line-height: 1.1; }
.yc-tile-sub { margin-top: 10px; font-size: 13.5px; color: var(--yc-mut); }
.yc-wallet { position: relative; overflow: hidden; border-radius: 24px; background: var(--yc-lime); color: var(--yc-on-lime); box-shadow: 0 1px 0 rgb(255 255 255 / 0.7) inset, 0 24px 48px -24px rgb(120 160 0 / 0.75); }
.yc-wallet::before { content: ''; position: absolute; inset: 0; background: linear-gradient(120deg, rgb(255 255 255 / 0.6), transparent 42%), radial-gradient(circle at 100% 100%, rgb(11 15 12 / 0.1), transparent 55%); pointer-events: none; }
.yc-wallet > * { position: relative; }
.yc-wallet .yc-tile-lab, .yc-wallet .yc-tile-sub { color: rgb(11 15 12 / 0.7); }
.yc-wallet .yc-tile-num { font-size: clamp(44px, 5.4vw, 68px); }
.yc-wallet-btn { margin-top: 20px; background: var(--yc-on-lime); color: var(--yc-lime); }
.yc-wallet-btn:hover { box-shadow: 0 10px 22px -8px rgb(11 15 12 / 0.65); transform: translateY(-2px); }
.yc-recent { margin-top: 16px; padding: 24px 26px; }
.yc-recent header { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 10px; }
.yc-recent h2 { font-size: 19px; letter-spacing: -0.015em; font-weight: 750; }
.yc-recent header a { font-size: 14px; color: var(--yc-mut); }
.yc-recent header a:hover { color: var(--yc-ink); }
.yc-row { width: 100%; display: grid; grid-template-columns: 120px 1fr auto 16px; gap: 14px; align-items: center; padding: 14px 0; border-top: 1px solid var(--yc-line); text-align: left; font-size: 15px; }
.yc-row-time { font-size: 13.5px; color: var(--yc-mut); }
.yc-row-model { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 14px; }
.yc-row i { font-style: normal; color: var(--yc-dim); transition: transform 0.15s ease; }
.yc-row[aria-expanded='true'] i { transform: rotate(90deg); }
.yc-detail { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px 20px; margin-bottom: 14px; padding: 16px 18px; border-radius: 14px; background: var(--yc-fill); }
.yc-detail dt { font-size: 12px; color: var(--yc-mut); }
.yc-detail dd { margin-top: 2px; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13px; overflow-wrap: anywhere; }
.yc-empty { padding: 28px 0 12px; border-top: 1px solid var(--yc-line); color: var(--yc-mut); }
.yc-tip { margin-top: 10px; font-size: 12.5px; color: var(--yc-dim); }
@media (max-width: 760px) {
  .yc-tiles { grid-template-columns: 1fr; }
  .yc-next { flex-wrap: wrap; }
  .yc-row { grid-template-columns: 1fr auto 16px; }
  .yc-row-time { grid-column: 1 / -1; margin-bottom: -10px; }
  .yc-detail { grid-template-columns: repeat(2, 1fr); }
}
</style>
