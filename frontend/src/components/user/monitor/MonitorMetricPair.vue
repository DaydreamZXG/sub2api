<template>
  <div class="mt-5 grid grid-cols-2 gap-2">
    <div
      class="ych-metric"
    >
      <div
        class="ych-metric-label"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="ych-metric-value">
        {{ primaryValue }}<span class="ych-metric-unit">{{ primaryUnit }}</span>
      </div>
    </div>
    <div
      class="ych-metric"
    >
      <div
        class="ych-metric-label"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="ych-metric-value">
        {{ secondaryValue }}<span class="ych-metric-unit">{{ secondaryUnit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  primaryLabel: string
  primaryValue: string
  primaryUnit: string
  primaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
  secondaryLabel: string
  secondaryValue: string
  secondaryUnit: string
  secondaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
}>()
</script>

<style scoped>
.ych-metric { padding: 12px; border-radius: 14px; background: var(--yc-fill); }
.ych-metric-label { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--yc-mut); }
.ych-metric-value { margin-top: 6px; font-size: 18px; font-weight: 720; color: var(--yc-ink); font-variant-numeric: tabular-nums; }
.ych-metric-unit { margin-left: 2px; font-size: 12px; font-weight: 400; color: var(--yc-mut); }
</style>
