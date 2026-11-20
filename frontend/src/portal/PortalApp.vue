<template>
  <div class="pt">
    <header class="pt-top">
      <div class="pt-brand">
        <img v-if="config.portal_logo" :src="config.portal_logo" alt="" />
        <b>{{ config.site_name }}</b>
      </div>
      <div class="pt-top-right">
        <button v-if="usage" type="button" class="pt-link" data-test="portal-switch" @click="signOut">{{ say('switchKey') }}</button>
        <button type="button" class="pt-link" data-test="portal-locale" @click="toggleLocale">{{ locale === 'zh' ? 'English' : '中文' }}</button>
      </div>
    </header>

    <main class="pt-main">
      <section v-if="!usage" class="pt-card pt-enter" data-test="portal-enter">
        <p class="pt-tagline">{{ say('tagline') }}</p>
        <h1>{{ say('enterTitle') }}</h1>
        <form @submit.prevent="lookUp(input)">
          <input v-model.trim="input" type="text" autocomplete="off" spellcheck="false" :placeholder="say('enterPlaceholder')" data-test="portal-key-input" />
          <button type="submit" class="pt-btn" :disabled="loading || !input" data-test="portal-look-up">{{ loading ? say('loading') : say('enter') }}</button>
        </form>
        <p v-if="error" class="pt-error" data-test="portal-error">{{ say(error) }}</p>
        <p class="pt-hint">{{ say('enterHint') }}</p>
      </section>

      <template v-else>
        <section class="pt-card pt-budget" data-test="portal-budget">
          <div class="pt-budget-head">
            <div>
              <p class="pt-label">{{ say('remaining') }}</p>
              <p class="pt-big" data-test="portal-remaining">{{ money(usage.quota?.remaining) }}</p>
            </div>
            <div class="pt-budget-side">
              <span class="pt-chip" :class="`is-${status}`" data-test="portal-status">{{ say(`status.${status}`) }}</span>
              <button type="button" class="pt-link" :disabled="loading" data-test="portal-refresh" @click="lookUp(key)">{{ say('refresh') }}</button>
            </div>
          </div>
          <div class="pt-bar"><i :style="{ width: `${usedPercent}%` }"></i></div>
          <p class="pt-sub">
            {{ say('used', { used: money(usage.quota?.used), limit: money(usage.quota?.limit) }) }}
            ·
            {{ expiryText }}
          </p>
          <p v-if="status !== 'active'" class="pt-error">{{ say('statusHelp') }}</p>
        </section>

        <div class="pt-two">
          <section class="pt-card">
            <p class="pt-label">{{ say('today') }}</p>
            <p class="pt-mid" data-test="portal-today">{{ money(usage.usage?.today?.actual_cost) }}</p>
            <p class="pt-sub">{{ say('requests', { n: count(usage.usage?.today?.requests) }) }} · {{ say('tokens', { n: count(usage.usage?.today?.total_tokens) }) }}</p>
          </section>
          <section class="pt-card">
            <p class="pt-label">{{ say('total') }}</p>
            <p class="pt-mid" data-test="portal-total">{{ money(usage.usage?.total?.actual_cost) }}</p>
            <p class="pt-sub">{{ say('requests', { n: count(usage.usage?.total?.requests) }) }} · {{ say('tokens', { n: count(usage.usage?.total?.total_tokens) }) }}</p>
          </section>
        </div>

        <section class="pt-card">
          <h2>{{ say('connect') }}</h2>
          <p class="pt-sub">{{ say('connectLead') }}</p>
          <div class="pt-copy">
            <span>{{ say('baseUrl') }}</span>
            <code data-test="portal-base-url">{{ origin }}</code>
            <button type="button" class="pt-link" @click="copy(origin, 'url')">{{ copiedWhat === 'url' ? say('copied') : say('copy') }}</button>
          </div>
          <div class="pt-copy">
            <span>{{ say('key') }}</span>
            <code>{{ maskedKey }}</code>
            <button type="button" class="pt-link" data-test="portal-copy-key" @click="copy(key, 'key')">{{ copiedWhat === 'key' ? say('copied') : say('copy') }}</button>
          </div>
          <ul class="pt-notes">
            <li>{{ say('claude') }}</li>
            <li>{{ say('openai') }}</li>
            <li>{{ say('others') }}</li>
          </ul>
          <template v-if="models.length">
            <p class="pt-label pt-models-label">{{ say('available') }}</p>
            <p class="pt-models" data-test="portal-models"><span v-for="m in models" :key="m">{{ m }}</span></p>
          </template>
        </section>

        <section class="pt-card">
          <h2>{{ say('daily', { n: DAYS }) }}</h2>
          <p v-if="daily.length === 0" class="pt-sub" data-test="portal-daily-empty">{{ say('dailyEmpty') }}</p>
          <table v-else class="pt-table" data-test="portal-daily">
            <thead>
              <tr><th>{{ say('date') }}</th><th>{{ say('calls') }}</th><th>{{ say('amount') }}</th><th></th></tr>
            </thead>
            <tbody>
              <tr v-for="day in daily" :key="day.date">
                <td>{{ day.date }}</td>
                <td>{{ count(day.requests) }}</td>
                <td>{{ money(day.actual_cost) }}</td>
                <td class="pt-spark"><i :style="{ width: `${dailyMax > 0 ? Math.round((day.actual_cost / dailyMax) * 100) : 0}%` }"></i></td>
              </tr>
            </tbody>
          </table>
        </section>

        <section v-if="byModel.length" class="pt-card">
          <h2>{{ say('models') }}</h2>
          <table class="pt-table" data-test="portal-by-model">
            <thead>
              <tr><th>{{ say('model') }}</th><th>{{ say('calls') }}</th><th>{{ say('amount') }}</th></tr>
            </thead>
            <tbody>
              <tr v-for="row in byModel" :key="row.model">
                <td>{{ row.model }}</td>
                <td>{{ count(row.requests) }}</td>
                <td>{{ money(row.actual_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </section>
      </template>

      <p v-if="config.contact_info" class="pt-contact" data-test="portal-contact"><b>{{ say('contact') }}</b> {{ config.contact_info }}</p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { messages, pickLocale, type PortalLocale } from './messages'
import type { PortalConfig } from './mount'

interface Figures {
  requests?: number
  total_tokens?: number
  actual_cost?: number
}
interface KeyUsage {
  status?: string
  quota?: { limit: number; used: number; remaining: number }
  expires_at?: string | null
  days_until_expiry?: number
  usage?: { today?: Figures; total?: Figures }
  daily_usage?: { date: string; requests: number; actual_cost: number }[]
  model_stats?: { model: string; requests: number; actual_cost: number }[]
}

const props = defineProps<{ config: PortalConfig }>()

const KEY_STORE = 'portal_key'
const DAYS = 30
const origin = window.location.origin

const locale = ref<PortalLocale>(pickLocale())
const input = ref('')
const key = ref('')
const loading = ref(false)
const error = ref('')
const usage = ref<KeyUsage | null>(null)
const models = ref<string[]>([])
const copiedWhat = ref('')

// Deliberately not named like the main site's translate function: its locale check takes every
// call of that name for a key of the site's own locale files.
function say(name: string, params: Record<string, string | number> = {}): string {
  const text = messages[locale.value][name] ?? name
  return text.replace(/\{(\w+)\}/g, (_, p: string) => String(params[p] ?? ''))
}

function toggleLocale() {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
  localStorage.setItem('portal_locale', locale.value)
}

const money = (value: number | null | undefined) => `${props.config.sign}${(value ?? 0).toFixed(2)}`
const count = (value: number | null | undefined) => (value ?? 0).toLocaleString('en-US')

const status = computed(() => {
  const value = usage.value?.status ?? 'inactive'
  return ['active', 'inactive', 'quota_exhausted', 'expired'].includes(value) ? value : 'inactive'
})
const usedPercent = computed(() => {
  const quota = usage.value?.quota
  return quota && quota.limit > 0 ? Math.min(Math.round((quota.used / quota.limit) * 100), 100) : 0
})
const expiryText = computed(() => {
  const at = usage.value?.expires_at
  if (!at) return say('noExpiry')
  const date = new Date(at).toLocaleDateString(locale.value === 'zh' ? 'zh-CN' : 'en-CA')
  const days = usage.value?.days_until_expiry ?? 0
  return days > 0 ? say('expires', { date, days }) : say('expiresToday', { date })
})
// Newest day first, and only days with something on them.
const daily = computed(() => [...(usage.value?.daily_usage ?? [])].filter((d) => d.requests > 0).reverse())
const dailyMax = computed(() => Math.max(0, ...daily.value.map((d) => d.actual_cost)))
const byModel = computed(() => [...(usage.value?.model_stats ?? [])].sort((a, b) => b.actual_cost - a.actual_cost))
const maskedKey = computed(() => (key.value.length > 14 ? `${key.value.slice(0, 8)}…${key.value.slice(-4)}` : key.value))

async function lookUp(value: string) {
  if (!value || loading.value) return
  loading.value = true
  error.value = ''
  try {
    const headers = { Authorization: `Bearer ${value}` }
    const res = await fetch(`/v1/usage?days=${DAYS}`, { headers })
    if (res.status === 401 || res.status === 403) {
      error.value = 'badKey'
      signOut(false)
      return
    }
    if (!res.ok) throw new Error(String(res.status))
    usage.value = (await res.json()) as KeyUsage
    key.value = value
    localStorage.setItem(KEY_STORE, value)
    void loadModels(headers)
  } catch {
    error.value = 'failed'
  } finally {
    loading.value = false
  }
}

async function loadModels(headers: Record<string, string>) {
  try {
    const res = await fetch('/v1/models', { headers })
    if (!res.ok) return
    const body = (await res.json()) as { data?: { id?: string }[] }
    models.value = (body.data ?? []).map((m) => m.id ?? '').filter(Boolean)
  } catch {
    // The list is a convenience; the page is complete without it.
  }
}

function signOut(clearError = true) {
  localStorage.removeItem(KEY_STORE)
  usage.value = null
  models.value = []
  key.value = ''
  input.value = ''
  if (clearError) error.value = ''
}

async function copy(text: string, what: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const area = document.createElement('textarea')
    area.value = text
    document.body.appendChild(area)
    area.select()
    document.execCommand('copy')
    area.remove()
  }
  copiedWhat.value = what
  window.setTimeout(() => {
    if (copiedWhat.value === what) copiedWhat.value = ''
  }, 1500)
}

onMounted(() => {
  const saved = localStorage.getItem(KEY_STORE)
  if (saved) {
    input.value = saved
    void lookUp(saved)
  }
})
</script>

<style scoped>
.pt { --pt-bg: #f6f7f9; --pt-paper: #fff; --pt-ink: #14171c; --pt-mut: #5b6470; --pt-line: rgb(20 23 28 / 0.1); --pt-fill: rgb(20 23 28 / 0.05); --pt-accent: #2f5bea; --pt-on-accent: #fff; --pt-warn: #b45309; min-height: 100vh; background: var(--pt-bg); color: var(--pt-ink); font-family: system-ui, -apple-system, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif; font-variant-numeric: tabular-nums; -webkit-font-smoothing: antialiased; }
:global(.dark) .pt { --pt-bg: #0e1116; --pt-paper: #161a21; --pt-ink: #eef1f5; --pt-mut: rgb(238 241 245 / 0.66); --pt-line: rgb(238 241 245 / 0.12); --pt-fill: rgb(238 241 245 / 0.07); --pt-accent: #7f9cff; --pt-on-accent: #0e1116; --pt-warn: #fbbf24; }
.pt-top { display: flex; align-items: center; justify-content: space-between; gap: 16px; max-width: 880px; margin: 0 auto; padding: 20px 20px 0; }
.pt-brand { display: flex; align-items: center; gap: 10px; min-width: 0; font-size: 18px; }
.pt-brand img { width: 28px; height: 28px; border-radius: 6px; object-fit: contain; }
.pt-brand b { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pt-top-right { display: flex; flex: none; gap: 16px; }
.pt-main { display: grid; gap: 16px; max-width: 880px; margin: 0 auto; padding: 20px 20px 56px; }
.pt-card { padding: 24px; border: 1px solid var(--pt-line); border-radius: 16px; background: var(--pt-paper); }
.pt-card h1 { margin: 6px 0 16px; font-size: 26px; font-weight: 750; letter-spacing: -0.01em; }
.pt-card h2 { font-size: 17px; font-weight: 700; }
.pt-tagline, .pt-label, .pt-sub, .pt-hint { color: var(--pt-mut); font-size: 14px; line-height: 1.6; }
.pt-label { font-size: 13px; }
.pt-sub { margin-top: 6px; }
.pt-hint { margin-top: 14px; font-size: 13px; }
.pt-error { margin-top: 12px; color: var(--pt-warn); font-size: 14px; }
.pt-enter form { display: flex; gap: 10px; }
.pt-enter input { flex: 1; min-width: 0; height: 46px; padding: 0 14px; border: 1px solid var(--pt-line); border-radius: 12px; background: var(--pt-paper); color: var(--pt-ink); font-size: 15px; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.pt-btn { flex: none; height: 46px; padding: 0 22px; border-radius: 12px; background: var(--pt-accent); color: var(--pt-on-accent); font-size: 15px; font-weight: 650; }
.pt-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.pt-link { flex: none; color: var(--pt-accent); font-size: 14px; white-space: nowrap; }
.pt-link:disabled { opacity: 0.5; }
.pt-budget-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.pt-budget-side { display: grid; justify-items: end; gap: 8px; }
.pt-big { margin-top: 2px; font-size: 40px; font-weight: 780; letter-spacing: -0.02em; }
.pt-mid { margin-top: 2px; font-size: 26px; font-weight: 740; }
.pt-chip { padding: 3px 12px; border-radius: 999px; background: var(--pt-fill); color: var(--pt-mut); font-size: 13px; font-weight: 600; white-space: nowrap; }
.pt-chip.is-active { background: color-mix(in srgb, #16a34a 16%, transparent); color: #15803d; }
:global(.dark) .pt-chip.is-active { color: #86efac; }
.pt-bar { height: 8px; margin-top: 16px; border-radius: 999px; background: var(--pt-fill); overflow: hidden; }
.pt-bar i { display: block; height: 100%; border-radius: 999px; background: var(--pt-accent); }
.pt-two { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
.pt-copy { display: flex; align-items: center; gap: 12px; margin-top: 12px; padding: 12px 14px; border-radius: 12px; background: var(--pt-fill); }
.pt-copy span { flex: none; width: 72px; color: var(--pt-mut); font-size: 13px; }
.pt-copy code { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px; }
.pt-notes { margin: 16px 0 0; padding-left: 20px; list-style: disc; color: var(--pt-mut); font-size: 14px; line-height: 1.7; }
.pt-notes li + li { margin-top: 4px; }
.pt-models-label { margin-top: 18px; }
.pt-models { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.pt-models span { padding: 3px 10px; border-radius: 8px; background: var(--pt-fill); font-size: 13px; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.pt-table { width: 100%; margin-top: 12px; border-collapse: collapse; font-size: 14px; }
.pt-table th { padding: 8px 0; color: var(--pt-mut); font-size: 13px; font-weight: 500; text-align: left; }
.pt-table td { padding: 10px 0; border-top: 1px solid var(--pt-line); overflow-wrap: anywhere; }
.pt-table th:nth-child(2), .pt-table td:nth-child(2), .pt-table th:nth-child(3), .pt-table td:nth-child(3) { padding-left: 12px; text-align: right; white-space: nowrap; }
.pt-spark { width: 30%; padding-left: 16px !important; }
.pt-spark i { display: block; height: 6px; min-width: 2px; border-radius: 999px; background: var(--pt-accent); opacity: 0.7; }
.pt-contact { color: var(--pt-mut); font-size: 14px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; }
.pt-contact b { margin-right: 8px; color: var(--pt-ink); font-weight: 600; }
@media (max-width: 640px) {
  .pt-card { padding: 20px 18px; }
  .pt-big { font-size: 34px; }
  .pt-two { grid-template-columns: minmax(0, 1fr); }
  .pt-enter form { flex-direction: column; }
  .pt-copy span { width: 56px; }
  .pt-spark { display: none; }
}
</style>
