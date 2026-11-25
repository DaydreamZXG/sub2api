<template>
  <div class="pz">
    <header>
      <h1 class="yc-page-h1">{{ t('yc.prices.title') }}</h1>
      <p class="yc-page-lead">{{ t('modelPlaza.lead') }}</p>
    </header>

    <!-- Price notes written by the admin (Markdown). -->
    <div v-if="descriptionHtml" class="yc-surface pz-notes" v-html="descriptionHtml"></div>

    <p v-if="!isAuthenticated" class="yc-note pz-hint">
      <Icon name="infoCircle" size="sm" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <p v-if="loading" class="yc-surface pz-state" role="status">{{ t('yc.prices.loading') }}</p>
    <p v-else-if="error" class="yc-surface pz-state pz-state-bad" role="alert">{{ t('modelPlaza.loadFailed') }}</p>
    <template v-else>
      <div class="pz-filter">
        <div class="pz-chips" role="group" :aria-label="t('modelPlaza.filters.modelLabel')">
          <button
            v-for="f in familyChips"
            :key="f.key"
            type="button"
            class="pz-chip"
            :class="{ 'is-on': family === f.key }"
            :aria-pressed="family === f.key"
            data-test="plaza-family"
            @click="family = f.key"
          >
            <ModelIcon v-if="f.sample" :model="f.sample" size="14px" />
            {{ f.label }}
          </button>
        </div>
        <input
          v-model.trim="search"
          type="search"
          class="yc-input pz-search"
          :placeholder="t('modelPlaza.filters.searchPlaceholder')"
          :aria-label="t('modelPlaza.filters.searchPlaceholder')"
        />
      </div>

      <section v-if="shown.length > 0" class="yc-surface pz-list">
        <div class="pz-head" aria-hidden="true">
          <span>{{ t('modelPlaza.table.model') }}</span><span>{{ t('modelPlaza.table.input') }}</span>
          <span>{{ t('modelPlaza.table.output') }}</span><span>{{ t('modelPlaza.list.cacheRead') }}</span>
          <span>{{ t('modelPlaza.list.vsOfficial') }}</span>
        </div>
        <template v-for="m in shown" :key="m.name">
          <button
            type="button"
            class="pz-row"
            :class="{ 'is-open': open === m.name }"
            :aria-expanded="open === m.name"
            data-test="plaza-model"
            @click="open = open === m.name ? '' : m.name"
          >
            <span class="pz-model">
              <ModelIcon :model="m.name" size="24px" />
              <span class="pz-model-text">
                <b class="yc-mono">{{ m.name }}</b>
                <small>{{ m.familyLabel }} · {{ t('modelPlaza.list.lines', { n: m.entries.length }) }}</small>
              </span>
            </span>
            <span class="pz-cell" :data-label="t('modelPlaza.table.input')">
              <b class="yc-num">{{ money(m.input) }}<small v-if="m.varies && m.input != null">{{ t('modelPlaza.list.from') }}</small></b>
              <s v-if="m.officialInput != null">{{ officialText(m.official, 'input_price') }}</s>
            </span>
            <span class="pz-cell" :data-label="t('modelPlaza.table.output')">
              <b class="yc-num">{{ money(m.output) }}<small v-if="m.varies && m.output != null">{{ t('modelPlaza.list.from') }}</small></b>
              <s v-if="m.officialOutput != null">{{ officialText(m.official, 'output_price') }}</s>
            </span>
            <span class="pz-cell" :data-label="t('modelPlaza.list.cacheRead')">
              <span class="yc-num">{{ money(m.cacheRead) }}</span>
            </span>
            <span class="pz-off">
              <span v-if="m.share != null" class="yc-badge" :class="{ 'yc-badge-lime': m.share < 1 }" data-test="plaza-model-share">
                {{ priceChangeLabel(t, m.share) || t('modelPlaza.table.standardPrice') }}
              </span>
              <span class="pz-caret" aria-hidden="true"></span>
            </span>
          </button>
          <div v-if="open === m.name" class="pz-detail">
            <PlazaGroupSection v-for="e in m.entries" :key="e.group.id" :group="{ ...e.group, models: [e.model] }" />
          </div>
        </template>
      </section>
      <p v-else class="yc-surface pz-state">
        {{ searchActive ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}
      </p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse, PlazaModel, PlazaOfficialPricing } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'
import { unitsPerDollar } from '@/composables/ycMoneyState'
import { officialUsdPerToken, usdText, type OfficialField } from '@/utils/pricing'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import { familiesOf, modelFamily } from '@/utils/modelFamily'
import { BILLING_MODE_TOKEN } from '@/utils/billingMode'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)

const PER_MILLION = 1_000_000

const family = ref('all')
const search = ref('')
const open = ref('')
const searchActive = computed(() => search.value !== '')

const descriptionHtml = computed(() => {
  const md = props.response?.description?.trim()
  if (!md) return ''
  return DOMPurify.sanitize(marked.parse(md) as string)
})

interface PlazaEntry {
  group: ModelPlazaGroup
  model: PlazaModel
  /** $ per million input tokens on this line, null when it is not billed per token. */
  input: number | null
}

/** One model across every line that sells it; the headline is the cheapest line anyone can use. */
interface PlazaModelRow {
  name: string
  familyKey: string
  familyLabel: string
  rank: number
  entries: PlazaEntry[]
  input: number | null
  output: number | null
  cacheRead: number | null
  official: PlazaOfficialPricing | null
  /** Official prices in $ per million tokens, the basis of the discount. */
  officialInput: number | null
  officialOutput: number | null
  /** Headline input price as a share of the official one, null without an official price. */
  share: number | null
  /** The headline is the standard price; long context, time of day or reasoning effort can raise it. */
  varies: boolean
}

function rateOf(g: ModelPlazaGroup): number {
  return g.user_rate_multiplier ?? g.rate_multiplier
}

function paid(value: number | null | undefined, g: ModelPlazaGroup): number | null {
  return value == null ? null : (value * rateOf(g) * PER_MILLION) / unitsPerDollar.value
}

function listed(o: PlazaOfficialPricing | null, field: OfficialField): number | null {
  const v = officialUsdPerToken(o, field)
  return v == null ? null : v * PER_MILLION
}

function varies(e: PlazaEntry): boolean {
  const { group: g, model: m } = e
  const p = m.pricing
  return (
    (g.long_context_pricing_enabled && (p?.intervals?.length ?? 0) > 1) ||
    (m.time_pricing?.periods ?? []).some((period) => period.multiplier > 1) ||
    (g.peak_rate_enabled && g.peak_rate_multiplier > 1) ||
    Object.values(p?.reasoning_effort_multipliers ?? {}).some((x) => x > 1)
  )
}

/** Priced lines first, then open pay-as-you-go ones, so the headline is a price anyone can get with any tool. */
function lineOrder(a: PlazaEntry, b: PlazaEntry): number {
  const unpriced = (e: PlazaEntry) => Number(e.input == null)
  const plan = (e: PlazaEntry) => Number(e.group.subscription_type === 'subscription')
  const app = (e: PlazaEntry) => Number(e.group.app_only === true)
  return unpriced(a) - unpriced(b) || plan(a) - plan(b) || app(a) - app(b) || (a.input ?? 0) - (b.input ?? 0)
}

/** Per-request and per-image models are priced per unit in the line table, not per token. */
function tokenBilled(m: PlazaModel): boolean {
  return (m.pricing?.billing_mode || BILLING_MODE_TOKEN) === BILLING_MODE_TOKEN
}

const models = computed<PlazaModelRow[]>(() => {
  const byName = new Map<string, PlazaEntry[]>()
  for (const group of props.response?.groups ?? []) {
    for (const model of group.models) {
      const entries = byName.get(model.name) ?? []
      entries.push({ group, model, input: tokenBilled(model) ? paid(model.pricing?.input_price, group) : null })
      byName.set(model.name, entries)
    }
  }
  const rows: PlazaModelRow[] = []
  for (const [name, entries] of byName) {
    entries.sort(lineOrder)
    const head = entries[0]
    const f = modelFamily(name)
    const official = entries.find((e) => e.model.official_pricing)?.model.official_pricing ?? null
    const officialInput = listed(official, 'input_price')
    rows.push({
      name,
      familyKey: f?.key ?? 'other',
      familyLabel: f?.label ?? t('modelPlaza.list.other'),
      rank: f?.rank ?? Number.MAX_SAFE_INTEGER,
      entries,
      input: head.input,
      output: head.input == null ? null : paid(head.model.pricing?.output_price, head.group),
      cacheRead: head.input == null ? null : paid(head.model.pricing?.cache_read_price, head.group),
      official,
      officialInput,
      officialOutput: listed(official, 'output_price'),
      share: head.input != null && officialInput ? Math.round((head.input / officialInput) * 1000) / 1000 : null,
      varies: varies(head)
    })
  }
  return rows.sort((a, b) => a.rank - b.rank || a.name.localeCompare(b.name))
})

const familyChips = computed(() => {
  const chips: { key: string; label: string; sample: string }[] = [
    { key: 'all', label: t('modelPlaza.filters.all'), sample: '' },
    ...familiesOf(models.value.map((m) => m.name)).map((f) => ({ key: f.key, label: f.label, sample: f.sample }))
  ]
  if (models.value.some((m) => m.familyKey === 'other')) {
    chips.push({ key: 'other', label: t('modelPlaza.list.other'), sample: '' })
  }
  return chips
})

const shown = computed(() => {
  const q = search.value.toLowerCase()
  return models.value.filter(
    (m) => (family.value === 'all' || m.familyKey === family.value) && (!q || m.name.toLowerCase().includes(q))
  )
})

function money(value: number | null): string {
  return value == null ? '-' : usdText(value)
}

function officialText(o: PlazaOfficialPricing | null, field: OfficialField): string {
  const v = officialUsdPerToken(o, field)
  return v == null ? '-' : usdText(v * PER_MILLION)
}
</script>

<style scoped>
.pz { display: grid; gap: 20px; }
.pz-hint { display: flex; align-items: center; gap: 6px; margin: 0; }
.pz-state { margin: 0; padding: 48px 24px; text-align: center; font-size: 14.5px; color: var(--yc-mut); }
.pz-state-bad { color: #c2410c; }
.pz-lines { display: grid; gap: 16px; }
.pz-notes { padding: 18px 24px; font-size: 14px; line-height: 1.7; overflow-wrap: anywhere; color: var(--yc-ink); }
.pz-notes :deep(h1), .pz-notes :deep(h2), .pz-notes :deep(h3) { margin: 12px 0 8px; font-size: 15px; font-weight: 700; }
.pz-notes :deep(:first-child) { margin-top: 0; }
.pz-notes :deep(p) { margin: 0 0 8px; }
.pz-notes :deep(:last-child) { margin-bottom: 0; }
.pz-notes :deep(a) { text-decoration: underline; text-underline-offset: 4px; }
.pz-notes :deep(ul) { margin: 0 0 8px; padding-left: 20px; list-style: disc; }
.pz-notes :deep(ol) { margin: 0 0 8px; padding-left: 20px; list-style: decimal; }
.pz-notes :deep(code) { padding: 2px 6px; border-radius: 6px; background: var(--yc-fill); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12.5px; }
.pz-notes :deep(blockquote) { margin: 8px 0; padding-left: 12px; border-left: 3px solid var(--yc-line); color: var(--yc-mut); }
.pz-filter { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; }
.pz-chips { display: flex; flex-wrap: wrap; gap: 8px; }
.pz-chip { display: inline-flex; align-items: center; gap: 6px; padding: 6px 14px; border: 1px solid var(--yc-line); border-radius: 999px; background: var(--yc-paper); font-size: 13px; font-weight: 600; color: var(--yc-mut); transition: background-color 0.15s ease, color 0.15s ease; }
.pz-chip:hover { color: var(--yc-ink); }
.pz-chip.is-on { border-color: transparent; background: var(--yc-on); color: #fff; }
.pz-search { max-width: 280px; }
.pz-list { padding: 6px 24px 10px; }
.pz-head, .pz-row { display: grid; grid-template-columns: minmax(0, 2fr) 1fr 1fr 0.8fr 120px; gap: 16px; align-items: center; }
.pz-head { padding: 14px 0 10px; font-size: 12.5px; font-weight: 650; color: var(--yc-mut); }
.pz-head span:last-child { text-align: right; }
.pz-row { width: 100%; padding: 14px 0; border-top: 1px solid var(--yc-line); text-align: left; }
.pz-row:hover b.yc-mono { text-decoration: underline; text-underline-offset: 4px; }
.pz-model { display: flex; align-items: center; gap: 12px; min-width: 0; }
.pz-model-text { display: grid; min-width: 0; }
.pz-model-text b { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14.5px; color: var(--yc-ink); }
.pz-model-text small { font-size: 12.5px; color: var(--yc-mut); }
.pz-cell { display: flex; flex-wrap: wrap; align-items: baseline; gap: 4px 8px; color: var(--yc-ink); }
.pz-cell b { font-size: 18px; }
.pz-cell b small { margin-left: 2px; font-size: 12px; font-weight: 500; color: var(--yc-mut); }
.pz-cell s { font-size: 13px; color: var(--yc-dim); }
.pz-off { display: flex; align-items: center; justify-content: flex-end; gap: 10px; }
.pz-caret { width: 8px; height: 8px; border-right: 2px solid var(--yc-dim); border-bottom: 2px solid var(--yc-dim); transform: rotate(45deg); transition: transform 0.2s ease; }
.pz-row.is-open .pz-caret { transform: rotate(-135deg); }
.pz-detail { display: grid; gap: 12px; padding: 4px 0 18px; }
@media (max-width: 760px) {
  .pz-list { padding: 4px 16px 8px; }
  .pz-head { display: none; }
  .pz-row { grid-template-columns: 1fr 1fr; gap: 8px 12px; }
  .pz-model { grid-column: 1 / -1; }
  .pz-off { grid-column: 1 / -1; justify-content: flex-start; }
  .pz-cell::before { content: attr(data-label); flex-basis: 100%; font-size: 12px; color: var(--yc-mut); }
  .pz-search { max-width: none; flex: 1 1 100%; }
}
</style>
