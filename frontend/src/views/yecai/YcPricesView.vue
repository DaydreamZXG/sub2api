<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.prices.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.prices.lead') }}</p>

    <input v-model.trim="search" type="search" class="yc-input yc-prices-search" :placeholder="t('yc.prices.search')" :aria-label="t('yc.prices.search')" />

    <section class="yc-surface yc-prices">
      <header class="yc-prices-head">
        <span>{{ t('yc.prices.model') }}</span><span>{{ t('yc.prices.input') }}</span><span>{{ t('yc.prices.output') }}</span><span></span>
      </header>
      <p v-if="loading" class="yc-prices-empty">{{ t('yc.prices.loading') }}</p>
      <p v-else-if="shown.length === 0" class="yc-prices-empty">{{ t(prices.length ? 'yc.prices.noMatch' : 'yc.prices.none') }}</p>
      <template v-for="m in shown" :key="m.name">
        <button type="button" class="yc-prices-row" :aria-expanded="open === m.name" @click="open = open === m.name ? '' : m.name">
          <span class="yc-mono yc-prices-name">{{ m.name }} <span v-if="m.appOnly" class="yc-badge">{{ t('yc.prices.appOnly') }}</span></span>
          <span class="yc-prices-cell" :data-label="t('yc.prices.input')">
            <b class="yc-num">¥{{ price(m.input) }}<small v-if="m.varies">{{ t('yc.prices.from') }}</small></b><s v-if="m.share && m.officialInput">¥{{ price(m.officialInput) }}</s>
            <em v-if="m.app" class="yc-prices-app">{{ t('yc.prices.appPrice', { price: price(m.app.input) }) }}<template v-if="m.app.varies">{{ t('yc.prices.from') }}</template></em>
          </span>
          <span class="yc-prices-cell" :data-label="t('yc.prices.output')">
            <b class="yc-num">¥{{ price(m.output) }}<small v-if="m.varies">{{ t('yc.prices.from') }}</small></b><s v-if="m.share && m.officialOutput">¥{{ price(m.officialOutput) }}</s>
            <em v-if="m.app" class="yc-prices-app">{{ t('yc.prices.appPrice', { price: price(m.app.output) }) }}<template v-if="m.app.varies">{{ t('yc.prices.from') }}</template></em>
          </span>
          <span class="yc-prices-off"><span v-if="m.share" class="yc-badge">{{ t('yc.price.off', { zhe: (m.share * 10).toFixed(1), pct: Math.round((1 - m.share) * 100) }) }}</span></span>
        </button>
        <div v-if="open === m.name" class="yc-prices-detail">
          <table>
            <thead>
              <tr>
                <th>{{ t('yc.prices.line') }}</th><th>{{ t('yc.prices.input') }}</th><th>{{ t('yc.prices.output') }}</th>
                <th>{{ t('yc.prices.cacheRead') }}</th><th>{{ t('yc.prices.multiplier') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="l in m.lines" :key="l.name">
                <td>{{ l.name }} <span v-if="l.plan" class="yc-badge">{{ t('yc.shell.plans') }}</span><span v-if="l.app" class="yc-badge">{{ t('yc.prices.appOnly') }}</span></td><td>¥{{ price(l.input) }}</td><td>¥{{ price(l.output) }}</td>
                <td>{{ l.cacheRead == null ? '-' : `¥${price(l.cacheRead)}` }}</td><td>×{{ l.multiplier }}</td>
              </tr>
            </tbody>
          </table>
          <p v-if="m.lines.some((l) => l.varies)" class="yc-note">{{ t('yc.prices.varies') }}</p>
          <p v-if="m.usdInput != null && m.usdOutput != null" class="yc-note">
            {{ t('yc.prices.official', { input: price(m.usdInput), output: price(m.usdOutput), rate: cnyPerUsd }) }}
          </p>
        </div>
      </template>
    </section>
    <p v-if="prices.some((m) => m.app || m.appOnly)" class="yc-note yc-prices-foot">{{ t('yc.prices.appNote') }}</p>
    <p class="yc-note yc-prices-foot">{{ t('yc.prices.foot') }}</p>
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import YcShell from '@/components/yecai/YcShell.vue'
import { useYcPrices } from '@/composables/useYcPrices'

const { t } = useI18n()
const { prices, load, cnyPerUsd } = useYcPrices()

const loading = ref(true)
const search = ref('')
const open = ref('')
const shown = computed(() => {
  const q = search.value.toLowerCase()
  return q ? prices.value.filter((m) => m.name.toLowerCase().includes(q)) : prices.value
})

function price(value: number): string {
  if (value >= 100) return value.toFixed(0)
  return value >= 10 ? value.toFixed(1) : value.toFixed(2)
}

onMounted(async () => {
  await load()
  loading.value = false
})
</script>

<style scoped>
.yc-prices-search { max-width: 360px; margin-top: 24px; }
.yc-prices { margin-top: 16px; padding: 8px 26px 12px; }
.yc-prices-head, .yc-prices-row { display: grid; grid-template-columns: minmax(0, 1.6fr) 1fr 1fr 96px; gap: 16px; align-items: center; }
.yc-prices-head { padding: 14px 0 10px; font-size: 12.5px; font-weight: 650; color: var(--yc-mut); }
.yc-prices-row { width: 100%; padding: 15px 0; border-top: 1px solid var(--yc-line); text-align: left; transition: background-color 0.15s ease; }
.yc-prices-row:hover .yc-prices-name { text-decoration: underline; text-underline-offset: 4px; }
.yc-prices-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14.5px; }
.yc-prices-cell { display: flex; align-items: baseline; gap: 8px; }
.yc-prices-cell b { font-size: 19px; }
.yc-prices-cell b small { font-size: 12px; font-weight: 500; color: var(--yc-mut); }
.yc-prices-cell s { font-size: 13px; color: var(--yc-dim); }
.yc-prices-cell { flex-wrap: wrap; }
.yc-prices-app { flex-basis: 100%; font-style: normal; font-size: 12.5px; font-weight: 600; color: var(--yc-ink); }
.yc-prices-app::before { content: ''; display: inline-block; width: 7px; height: 7px; margin-right: 6px; border-radius: 50%; background: var(--yc-lime); }
.yc-prices-off { text-align: right; }
.yc-prices-empty { padding: 28px 0; border-top: 1px solid var(--yc-line); color: var(--yc-mut); }
.yc-prices-detail { margin-bottom: 14px; padding: 14px 18px; border-radius: 14px; background: var(--yc-fill); overflow-x: auto; }
.yc-prices-detail table { width: 100%; border-collapse: collapse; font-size: 13.5px; }
.yc-prices-detail th { padding: 4px 12px 8px 0; text-align: left; font-size: 12px; font-weight: 600; color: var(--yc-mut); white-space: nowrap; }
.yc-prices-detail td { padding: 5px 12px 5px 0; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; white-space: nowrap; }
.yc-prices-detail .yc-note { margin-top: 10px; }
.yc-prices-foot { margin-top: 14px; }
@media (max-width: 760px) {
  .yc-prices { padding: 4px 18px 8px; }
  .yc-prices-head { display: none; }
  .yc-prices-row { grid-template-columns: 1fr 1fr; gap: 6px 12px; }
  .yc-prices-name { grid-column: 1 / 2; }
  .yc-prices-off { grid-column: 2 / 3; grid-row: 1; }
  .yc-prices-cell::before { content: attr(data-label); font-size: 12px; color: var(--yc-mut); }
}
</style>
