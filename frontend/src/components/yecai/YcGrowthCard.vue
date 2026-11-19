<template>
  <section v-if="growth && level" class="yc-surface yc-growth">
    <YcPlant class="yc-growth-plant" :stage="stage" :label="level.name" />

    <div class="yc-growth-main">
      <p class="yc-growth-lab">{{ t('yc.growth.title') }}</p>
      <p class="yc-growth-name">
        {{ level.name }}
        <span v-if="level.discount > 0" class="yc-badge yc-badge-lime">{{ t('yc.growth.off', { percent: percent(level.discount) }) }}</span>
      </p>
      <div
        class="yc-growth-bar"
        role="progressbar"
        aria-valuemin="0"
        aria-valuemax="100"
        :aria-valuenow="Math.round(progress * 100)"
      >
        <i :style="{ width: `${progress * 100}%` }" />
      </div>
      <p class="yc-growth-next">
        <template v-if="nextLevel">
          {{ t('yc.growth.next', { amount: yuan(nextLevel.threshold - growth.total_recharged), name: nextLevel.name, percent: percent(nextLevel.discount) }) }}
        </template>
        <template v-else>{{ t('yc.growth.top') }}</template>
      </p>
      <p class="yc-growth-note">{{ t(growth.manual_price ? 'yc.growth.manual' : 'yc.growth.note') }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getMyGrowth, type MyGrowth } from '@/api/growth'
import { useYcMoney } from '@/composables/useYcMoney'
import YcPlant from './YcPlant.vue'

// The plant has five drawings; a ladder of any length is spread across them.
const { t } = useI18n()
const { yuan } = useYcMoney()

const growth = ref<MyGrowth | null>(null)

const level = computed(() => (growth.value?.enabled ? growth.value.levels[growth.value.level_index] : undefined))
const nextLevel = computed(() => growth.value?.levels[growth.value.level_index + 1])
const stage = computed(() => {
  const last = (growth.value?.levels.length ?? 1) - 1
  return last <= 0 ? 0 : Math.round(((growth.value?.level_index ?? 0) / last) * 4)
})
const progress = computed(() => {
  if (!growth.value || !level.value || !nextLevel.value) return 1
  const span = nextLevel.value.threshold - level.value.threshold
  return Math.min(1, Math.max(0, (growth.value.total_recharged - level.value.threshold) / span))
})

function percent(discount: number): string {
  return String(Math.round(discount * 1000) / 10)
}

onMounted(async () => {
  try {
    growth.value = await getMyGrowth()
  } catch (error) {
    console.error('Failed to load the growth level:', error)
  }
})
</script>

<style scoped>
.yc-growth { display: flex; align-items: center; gap: 24px; margin-top: 16px; padding: 22px 26px; }
.yc-growth-plant { flex: none; width: 96px; height: 96px; }
.yc-growth-main { flex: 1; min-width: 0; }
.yc-growth-lab { font-size: 13px; font-weight: 650; color: var(--yc-mut); }
.yc-growth-name { display: flex; align-items: center; gap: 10px; margin-top: 4px; font-size: 26px; line-height: 1.2; letter-spacing: -0.02em; font-weight: 800; }
.yc-growth-bar { height: 8px; margin-top: 14px; overflow: hidden; border-radius: 999px; background: var(--yc-fill); }
.yc-growth-bar i { display: block; height: 100%; border-radius: inherit; background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-ink) inset; transition: width 0.6s cubic-bezier(0.2, 0.7, 0.2, 1); }
.yc-growth-next { margin-top: 10px; font-size: 14.5px; font-weight: 600; }
.yc-growth-note { margin-top: 4px; font-size: 12.5px; color: var(--yc-dim); }
@media (max-width: 760px) {
  .yc-growth { align-items: flex-start; gap: 16px; padding: 20px; }
  .yc-growth-plant { width: 72px; height: 72px; }
}
@media (prefers-reduced-motion: reduce) {
  .yc-growth-bar i { transition: none; }
}
</style>
