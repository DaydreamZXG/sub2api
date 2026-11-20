<template>
  <section class="ych-hero" :class="'ych-hero-' + overallStatus">
    <div class="ych-state">
      <span class="ych-badge" aria-hidden="true">
        <Icon :name="overallStatus === 'operational' ? 'check' : 'exclamationTriangle'" size="md" />
      </span>
      <div class="ych-state-text">
        <h2 data-test="mon-headline">{{ t(`channelStatus.headline.${overallStatus}`) }}</h2>
        <p>{{ t(`channelStatus.headlineHint.${overallStatus}`) }}</p>
      </div>
    </div>

    <div class="ych-tools">
      <div role="tablist" class="ych-tabs">
        <button
          v-for="opt in windowOptions"
          :key="opt.value"
          type="button"
          role="tab"
          :aria-selected="window === opt.value"
          class="ych-tab"
          :class="{ 'ych-tab-on': window === opt.value }"
          @click="emit('update:window', opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <button
        type="button"
        class="ych-refresh"
        :disabled="loading"
        :title="t('common.refresh')"
        :aria-label="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>

      <AutoRefreshButton
        v-if="autoRefresh"
        :enabled="autoRefresh.enabled.value"
        :interval-seconds="autoRefresh.intervalSeconds.value"
        :countdown="autoRefresh.countdown.value"
        :intervals="autoRefresh.intervals"
        @update:enabled="autoRefresh.setEnabled"
        @update:interval="autoRefresh.setInterval"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AutoRefreshButton from '@/components/common/AutoRefreshButton.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

defineProps<{
  overallStatus: OverallStatus
  intervalSeconds: number
  window: MonitorWindow
  loading: boolean
  autoRefresh?: {
    enabled: { value: boolean }
    intervalSeconds: { value: number }
    countdown: { value: number }
    intervals: readonly number[]
    setEnabled: (v: boolean) => void
    setInterval: (v: number) => void
  }
}>()

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])
</script>

<style scoped>
.ych-hero { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 16px 24px; margin-bottom: 20px; padding: 20px 24px; border: 1px solid var(--yc-line); border-radius: 22px; background: var(--yc-paper); }
.ych-state { display: flex; align-items: center; gap: 14px; min-width: 0; }
.ych-badge { display: grid; place-items: center; flex-shrink: 0; width: 44px; height: 44px; border-radius: 50%; background: var(--yc-lime); color: var(--yc-on-lime); }
.ych-hero-degraded .ych-badge { background: #f59e0b; color: #1c1300; }
.ych-state-text h2 { font-size: 20px; font-weight: 760; line-height: 1.25; color: var(--yc-ink); }
.ych-state-text p { margin-top: 2px; font-size: 13px; color: var(--yc-mut); }
.ych-tools { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
.ych-tabs { display: inline-flex; padding: 3px; border-radius: 999px; background: var(--yc-fill); }
.ych-tab { padding: 5px 14px; border-radius: 999px; font-size: 13px; color: var(--yc-mut); }
.ych-tab:hover { color: var(--yc-ink); }
.ych-tab-on { background: var(--yc-ink); color: var(--yc-paper); font-weight: 650; }
.ych-tab-on:hover { color: var(--yc-paper); }
.ych-refresh { display: grid; place-items: center; width: 34px; height: 34px; border: 1px solid var(--yc-line); border-radius: 50%; color: var(--yc-ink); }
.ych-refresh:not(:disabled):hover { background: var(--yc-fill); }
.ych-refresh:disabled { opacity: 0.5; cursor: not-allowed; }
@media (max-width: 640px) {
  .ych-hero { padding: 18px; }
}
</style>
