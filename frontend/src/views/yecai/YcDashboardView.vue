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

    <section class="yc-surface yc-mates">
      <header>
        <h2>{{ t('yc.dash.mates') }}</h2>
        <a :href="YC_CLIENT_PAGE">{{ t('yc.dash.matesGo') }} →</a>
      </header>
      <div class="yc-mate-row">
        <div v-for="m in MATES" :key="m.brand" class="yc-mate" :style="{ '--c': m.color }">
          <i :style="{ backgroundImage: `url(${roster})`, backgroundPosition: `${m.x}% 15%` }" aria-hidden="true"></i>
          <b>{{ t(m.name) }}</b>
          <small>{{ m.brand }}</small>
        </div>
      </div>
    </section>

    <YcGrowthCard />

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
import YcGrowthCard from '@/components/yecai/YcGrowthCard.vue'
import { useYcMoney } from '@/composables/useYcMoney'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import { YC_CLIENT_PAGE } from '@/utils/ycClient'
import roster from '@/assets/yecai/yeschoy-roster-showcase.webp'

// Horizontal position of each classmate's face in the roster artwork.
const MATES = [
  { name: 'yc.dash.mateJuzi', brand: 'Claude', color: '#ea7a2c', x: 5.4 },
  { name: 'yc.dash.mateAqing', brand: 'GPT', color: '#16a36a', x: 27.8 },
  { name: 'yc.dash.mateXiaojing', brand: 'DeepSeek', color: '#2f8fe0', x: 49.7 },
  { name: 'yc.dash.mateYueyue', brand: 'Kimi', color: '#4d7fe0', x: 71.9 },
  { name: 'yc.dash.mateDoubao', brand: 'Doubao', color: '#e0679b', x: 92.7 }
]

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
    { label: 'yc.dash.dMultiplier', value: priceChangeLabel(t, log.rate_multiplier) || t('modelPlaza.table.standardPrice') }
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
.yc-dash-h1 { font-size: clamp(26px, 3.4vw, 34px); line-height: 1.2; letter-spacing: -0.025em; font-weight: 650; }
.yc-next { display: flex; align-items: center; gap: 12px; margin-top: 20px; padding: 12px 14px 12px 16px; border-radius: 12px; background: var(--yc-paper); border: 1px solid var(--yc-line); color: var(--yc-ink); font-size: 14.5px; transition: border-color 0.15s ease; }
.yc-next:hover { border-color: var(--yc-dim); }
.yc-next-tag { flex: none; padding: 3px 9px; border-radius: 999px; background: var(--yc-accent-soft); color: var(--yc-accent); font-size: 12px; font-weight: 600; }
.yc-next-text { flex: 1; font-weight: 500; }
.yc-next-go { flex: none; padding: 6px 12px; border-radius: 8px; background: var(--yc-lime); color: var(--yc-on-lime); font-size: 13px; font-weight: 600; }
.yc-tiles { display: grid; grid-template-columns: 1.3fr 1fr 1fr; gap: 16px; margin-top: 16px; }
.yc-tile { padding: 22px 24px; }
.yc-tile-lab { font-size: 13px; font-weight: 500; color: var(--yc-mut); }
.yc-tile-num { margin-top: 8px; font-size: clamp(28px, 3vw, 34px); line-height: 1.15; }
.yc-tile-sub { margin-top: 6px; font-size: 13px; color: var(--yc-mut); }
.yc-wallet { border-radius: 12px; border: 1px solid color-mix(in srgb, var(--yc-accent) 22%, var(--yc-line)); background: linear-gradient(180deg, var(--yc-paper), color-mix(in srgb, var(--yc-accent) 5%, var(--yc-paper))); }
.yc-wallet .yc-tile-num { font-size: clamp(36px, 4vw, 48px); }
.yc-wallet-btn { margin-top: 16px; background: var(--yc-lime); color: var(--yc-on-lime); }
.yc-wallet-btn:hover { opacity: 0.86; }
.yc-mates { margin-top: 16px; padding: 22px 24px; }
.yc-mates header, .yc-recent header { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.yc-mates h2 { font-size: 17px; letter-spacing: -0.01em; font-weight: 600; }
.yc-mates header a, .yc-recent header a { font-size: 13.5px; font-weight: 500; color: var(--yc-accent); }
.yc-mate-row { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; }
.yc-mate { display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 16px 6px 14px; border: 1px solid var(--yc-line); border-radius: 10px; font-size: 14px; }
.yc-mate i { width: 44px; height: 44px; margin-bottom: 8px; border-radius: 50%; background-size: 1000% auto; background-repeat: no-repeat; box-shadow: 0 0 0 1px var(--yc-line); }
.yc-mate b { font-weight: 600; }
.yc-mate small { font-size: 12px; font-weight: 500; color: var(--c); }
.yc-recent { margin-top: 16px; padding: 22px 24px; }
.yc-recent header { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 10px; }
.yc-recent h2 { font-size: 17px; letter-spacing: -0.01em; font-weight: 600; }
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
  .yc-tiles { grid-template-columns: 1fr 1fr; }
  .yc-wallet { grid-column: 1 / -1; }
  .yc-mate-row { gap: 4px; }
  .yc-mate { font-size: 13px; }
  .yc-mate i { width: 40px; height: 40px; }
  .yc-mate small { font-size: 11px; }
  .yc-mate { padding: 12px 2px; border: 0; }
  .yc-next { flex-wrap: wrap; }
  .yc-row { grid-template-columns: 1fr auto 16px; }
  .yc-row-time { grid-column: 1 / -1; margin-bottom: -10px; }
  .yc-detail { grid-template-columns: repeat(2, 1fr); }
}
</style>
