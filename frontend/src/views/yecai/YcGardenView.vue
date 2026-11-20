<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.garden.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.garden.lead', { hours: garden?.ripen_hours ?? 24 }) }}</p>

    <p v-if="failed" class="yc-garden-empty">{{ t('yc.garden.failed') }}</p>
    <template v-else-if="garden">
      <YcField :class="{ 'has-yard': garden.enabled }">
        <template v-if="!growing.length" #sign>
          <p class="yc-garden-none-h">{{ t('yc.garden.emptyTitle') }}</p>
          <p class="yc-garden-none-p">{{ t(garden.enabled ? 'yc.garden.empty' : 'yc.garden.off') }}</p>
          <router-link v-if="garden.enabled" to="/topup" class="yc-btn yc-btn-sm">{{ t('yc.shell.topUp') }} →</router-link>
        </template>
        <article v-for="crop in growing" :key="crop.id" class="yc-plot" :class="{ ripe: isRipe(crop) }">
          <span v-if="crop.source !== 'order'" class="yc-badge yc-badge-lime">{{ t(`yc.garden.source.${crop.source}`) }}</span>
          <div class="yc-plot-bed">
            <YcPlant class="yc-plot-plant" :stage="stageOf(crop)" :label="t(isRipe(crop) ? 'yc.garden.ripe' : 'yc.garden.growing')" />
          </div>
          <p class="yc-plot-amount yc-num">{{ yuan(crop.amount - crop.stolen_amount) }}</p>
          <p v-if="crop.stolen_amount > 0" class="yc-plot-state">{{ t('yc.garden.stolen', { amount: yuan(crop.stolen_amount) }) }}</p>
          <p class="yc-plot-state">
            <template v-if="isRipe(crop)">{{ t('yc.garden.ripe') }}</template>
            <template v-else>{{ t('yc.garden.ripeIn', { time: remaining(crop) }) }}</template>
          </p>
          <div v-if="!isRipe(crop)" class="yc-plot-bar" role="progressbar" aria-valuemin="0" aria-valuemax="100" :aria-valuenow="Math.round(progressOf(crop) * 100)">
            <i :style="{ width: `${progressOf(crop) * 100}%` }" />
          </div>
          <button v-else type="button" class="yc-btn yc-btn-lime yc-btn-sm" :disabled="harvesting === crop.id" @click="harvest(crop)">
            {{ t('yc.garden.harvest') }}
          </button>
        </article>
        <div v-for="n in emptyPlots" :key="`empty-${n}`" class="yc-plot empty">{{ t('yc.garden.emptyPlot') }}</div>
      </YcField>

      <YcGardenSocial v-if="garden.enabled" :garden="garden" :now="now" @changed="refresh" />

      <section v-if="harvested.length" class="yc-surface yc-garden-log">
        <h2>{{ t('yc.garden.harvested') }}</h2>
        <ul>
          <li v-for="crop in harvested" :key="crop.id">
            <span>{{ day(crop.harvested_at) }}</span>
            <span class="yc-num">+{{ yuan(crop.harvested_amount) }}</span>
          </li>
        </ul>
      </section>
    </template>
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getMyGarden, harvestCrop, type GardenCrop, type MyGarden } from '@/api/garden'
import YcField from '@/components/yecai/YcField.vue'
import YcGardenSocial from '@/components/yecai/YcGardenSocial.vue'
import YcPlant from '@/components/yecai/YcPlant.vue'
import YcShell from '@/components/yecai/YcShell.vue'
import { useYcMoney } from '@/composables/useYcMoney'
import { useAppStore, useAuthStore } from '@/stores'

const { t, locale } = useI18n()
const { yuan } = useYcMoney()
const appStore = useAppStore()
const authStore = useAuthStore()

const garden = ref<MyGarden | null>(null)
const failed = ref(false)
const harvesting = ref<number | null>(null)
// Server time when loaded, advanced by a local ticker, so countdowns ignore the visitor's clock.
const now = ref(0)
let offset = 0
let ticker: ReturnType<typeof setInterval> | undefined

const growing = computed(() => (garden.value?.crops ?? []).filter((crop) => crop.status === 'growing'))
// The field always shows at least a row of beds, so an almost empty garden still reads as a garden.
const emptyPlots = computed(() => Math.max(0, 4 - growing.value.length))
const harvested = computed(() => (garden.value?.crops ?? []).filter((crop) => crop.status === 'harvested'))

const isRipe = (crop: GardenCrop) => new Date(crop.ripe_at).getTime() <= now.value
function progressOf(crop: GardenCrop): number {
  const planted = new Date(crop.planted_at).getTime()
  const span = new Date(crop.ripe_at).getTime() - planted
  return span <= 0 ? 1 : Math.min(1, Math.max(0, (now.value - planted) / span))
}
const stageOf = (crop: GardenCrop) => (isRipe(crop) ? 4 : Math.min(3, Math.floor(progressOf(crop) * 4)))

function remaining(crop: GardenCrop): string {
  const minutes = Math.max(1, Math.ceil((new Date(crop.ripe_at).getTime() - now.value) / 60000))
  const hours = Math.floor(minutes / 60)
  return hours > 0
    ? t('yc.garden.hoursMinutes', { hours, minutes: minutes % 60 })
    : t('yc.garden.minutes', { minutes })
}

function day(value?: string): string {
  return value ? new Date(value).toLocaleDateString(locale.value, { month: 'short', day: 'numeric' }) : ''
}

async function load() {
  try {
    const data = await getMyGarden()
    offset = new Date(data.now).getTime() - Date.now()
    now.value = Date.now() + offset
    garden.value = data
    failed.value = false
  } catch (error) {
    console.error('Failed to load the garden:', error)
    failed.value = true
  }
}

// Something a neighbour-side action changed: tries left, the dog, or the balance after a steal.
const refresh = () => Promise.all([load(), authStore.refreshUser()])

async function harvest(crop: GardenCrop) {
  harvesting.value = crop.id
  try {
    const { amount } = await harvestCrop(crop.id)
    appStore.showSuccess(t('yc.garden.harvestDone', { amount: yuan(amount) }))
    await refresh()
  } catch (error) {
    console.error('Failed to harvest:', error)
    appStore.showError(t('yc.garden.harvestFailed'))
    await load()
  } finally {
    harvesting.value = null
  }
}

onMounted(() => {
  void load()
  ticker = setInterval(() => (now.value = Date.now() + offset), 30_000)
})
onBeforeUnmount(() => clearInterval(ticker))
</script>

<style scoped>
.yc-garden-empty { margin-top: 28px; color: var(--yc-mut); }
.yc-garden-none-h { font-size: 20px; font-weight: 800; letter-spacing: -0.01em; }
.yc-garden-none-p { max-width: 460px; margin: 4px 0 10px; font-size: 14px; line-height: 1.5; }
.yc-garden-log { margin-top: 16px; padding: 22px 26px; }
.yc-garden-log h2 { font-size: 17px; font-weight: 800; }
.yc-garden-log ul { margin-top: 10px; }
.yc-garden-log li { display: flex; justify-content: space-between; padding: 10px 0; border-top: 1px solid var(--yc-line); font-size: 14.5px; }
.yc-garden-log li span:first-child { color: var(--yc-mut); }
</style>
