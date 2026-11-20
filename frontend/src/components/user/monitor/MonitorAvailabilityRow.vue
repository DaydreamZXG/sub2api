<template>
  <div class="mt-3 flex items-end justify-between">
    <div class="ych-avail-label">
      {{ windowLabel }}
    </div>
    <div class="flex items-baseline gap-0.5">
      <span
        class="text-3xl font-bold tabular-nums leading-none"
        :style="colorStyle"
      >
        {{ displayValue }}
      </span>
      <span
        class="text-base font-semibold leading-none"
        :style="colorStyle"
      >%</span>
    </div>
  </div>
  <div
    v-if="samplesLabel"
    class="ych-avail-note"
  >
    {{ samplesLabel }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { hslForPct } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  windowLabel: string
  value: number | null
  samplesLabel?: string
}>()

const { t } = useI18n()

const displayValue = computed(() => {
  if (props.value === null || Number.isNaN(props.value)) return t('monitorCommon.latencyEmpty')
  return props.value.toFixed(2)
})

const colorStyle = computed(() => {
  const colour = hslForPct(props.value)
  return colour ? { color: colour } : { color: 'rgb(156 163 175)' }
})
</script>

<style scoped>
.ych-avail-label { font-size: 12px; color: var(--yc-mut); }
.ych-avail-note { margin-top: 4px; font-size: 12px; color: var(--yc-mut); text-align: right; }
</style>
