<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.bills.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.bills.lead') }}</p>

    <div class="yc-opts yc-bills-range" role="group" :aria-label="t('yc.bills.range')">
      <button v-for="r in RANGES" :key="r.days" type="button" class="yc-opt" :class="{ 'is-on': r.days === days }" @click="pick(r.days)">
        {{ t(r.label) }}
      </button>
    </div>

    <section class="yc-bills-tiles">
      <div class="yc-surface yc-bills-tile">
        <p>{{ t('yc.bills.spent') }}</p>
        <b class="yc-num"><span class="yc-mark">{{ yuan(stats?.total_actual_cost) }}</span></b>
      </div>
      <div class="yc-surface yc-bills-tile">
        <p>{{ t('yc.bills.requests') }}</p>
        <b class="yc-num">{{ count(stats?.total_requests) }}</b>
      </div>
      <div class="yc-surface yc-bills-tile">
        <p>{{ t('yc.bills.tokens') }}</p>
        <b class="yc-num">{{ count(stats?.total_tokens) }}</b>
      </div>
    </section>

    <section class="yc-surface yc-bills-list">
      <p v-if="!loading && logs.length === 0" class="yc-bills-empty">{{ t('yc.bills.empty') }}</p>
      <template v-for="log in logs" :key="log.id">
        <button type="button" class="yc-bills-row" :aria-expanded="open === log.id" @click="open = open === log.id ? null : log.id">
          <span class="yc-bills-time">{{ formatDateTime(log.created_at) }}</span>
          <span class="yc-mono yc-bills-model">{{ log.model }}</span>
          <b>{{ yuan(log.actual_cost, 4) }}</b>
          <i aria-hidden="true">›</i>
        </button>
        <dl v-if="open === log.id" class="yc-bills-detail">
          <div v-for="d in details(log)" :key="d.label"><dt>{{ t(d.label) }}</dt><dd>{{ d.value }}</dd></div>
        </dl>
      </template>
      <button v-if="logs.length < total" type="button" class="yc-btn yc-btn-ghost yc-btn-sm yc-bills-more" :disabled="loading" @click="loadPage">
        {{ t('yc.bills.more', { n: total - logs.length }) }}
      </button>
    </section>
    <p class="yc-note yc-bills-foot">
      {{ t('yc.dash.expandTip') }}
      <router-link to="/usage" class="yc-link">{{ t('yc.bills.advanced') }} →</router-link>
    </p>
  </YcShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api/usage'
import type { UsageLog, UsageStatsResponse } from '@/types'
import { formatDateLocalInput, formatDateTime } from '@/utils/format'
import YcShell from '@/components/yecai/YcShell.vue'
import { useYcMoney } from '@/composables/useYcMoney'

const RANGES = [
  { days: 1, label: 'yc.bills.today' },
  { days: 7, label: 'yc.bills.week' },
  { days: 30, label: 'yc.bills.month' }
]
const PAGE_SIZE = 20

const { t } = useI18n()
const { yuan } = useYcMoney()

const days = ref(7)
const stats = ref<UsageStatsResponse | null>(null)
const logs = ref<UsageLog[]>([])
const total = ref(0)
const loading = ref(true)
const open = ref<number | null>(null)

function span() {
  return {
    start_date: formatDateLocalInput(new Date(Date.now() - (days.value - 1) * 86400000)),
    end_date: formatDateLocalInput(new Date())
  }
}

async function loadPage() {
  loading.value = true
  try {
    const page = await usageAPI.query({ ...span(), page: Math.floor(logs.value.length / PAGE_SIZE) + 1, page_size: PAGE_SIZE })
    logs.value = [...logs.value, ...page.items]
    total.value = page.total
  } catch (error) {
    console.error('Failed to load bills:', error)
  } finally {
    loading.value = false
  }
}

async function pick(next: number) {
  days.value = next
  logs.value = []
  total.value = 0
  open.value = null
  const { start_date, end_date } = span()
  usageAPI
    .getStatsByDateRange(start_date, end_date)
    .then((s) => (stats.value = s))
    .catch(() => (stats.value = null))
  await loadPage()
}

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

onMounted(() => pick(days.value))
</script>

<style scoped>
.yc-bills-range { margin-top: 24px; }
.yc-bills-tiles { display: grid; grid-template-columns: 1.3fr 1fr 1fr; gap: 16px; margin-top: 16px; }
.yc-bills-tile { padding: 22px 26px 24px; }
.yc-bills-tile p { font-size: 13px; font-weight: 650; color: var(--yc-mut); }
.yc-bills-tile b { display: block; margin-top: 8px; font-size: clamp(28px, 3.4vw, 42px); line-height: 1.15; overflow-wrap: anywhere; }
.yc-bills-list { margin-top: 16px; padding: 8px 26px 18px; }
.yc-bills-row { width: 100%; display: grid; grid-template-columns: 170px 1fr auto 16px; gap: 14px; align-items: center; padding: 14px 0; border-top: 1px solid var(--yc-line); text-align: left; font-size: 15px; }
.yc-bills-row:first-child { border-top: 0; }
.yc-bills-time { font-size: 13.5px; color: var(--yc-mut); }
.yc-bills-model { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px; }
.yc-bills-row i { font-style: normal; color: var(--yc-dim); transition: transform 0.15s ease; }
.yc-bills-row[aria-expanded='true'] i { transform: rotate(90deg); }
.yc-bills-detail { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px 20px; margin-bottom: 14px; padding: 16px 18px; border-radius: 14px; background: var(--yc-fill); }
.yc-bills-detail dt { font-size: 12px; color: var(--yc-mut); }
.yc-bills-detail dd { margin-top: 2px; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13px; overflow-wrap: anywhere; }
.yc-bills-empty { padding: 24px 0 10px; color: var(--yc-mut); }
.yc-bills-more { margin-top: 14px; }
.yc-bills-foot { margin-top: 14px; }
@media (max-width: 760px) {
  .yc-bills-tiles { grid-template-columns: 1fr; }
  .yc-bills-list { padding: 4px 18px 14px; }
  .yc-bills-row { grid-template-columns: 1fr auto 16px; }
  .yc-bills-time { grid-column: 1 / -1; margin-bottom: -10px; }
  .yc-bills-detail { grid-template-columns: repeat(2, 1fr); }
}
</style>
