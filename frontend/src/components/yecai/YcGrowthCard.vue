<template>
  <section v-if="growth && level" class="yc-surface yc-growth">
    <svg class="yc-growth-plant" viewBox="10 14 76 76" role="img" :aria-label="level.name">
      <ellipse cx="48" cy="80" rx="30" ry="7" class="soil" />
      <template v-if="stage === 0">
        <ellipse cx="48" cy="74" rx="6" ry="4.5" class="leaf" />
      </template>
      <template v-else>
        <path :d="STEMS[stage]" class="stem" />
        <path v-for="(d, i) in LEAVES[stage]" :key="i" :d="d" class="leaf" />
        <g v-if="stage >= 3">
          <circle v-for="(c, i) in BLOOMS[stage]" :key="i" :cx="c[0]" :cy="c[1]" :r="c[2]" class="bloom" />
        </g>
      </template>
    </svg>

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

// The plant has five drawings; a ladder of any length is spread across them.
const STEMS = ['', 'M48 76V62', 'M48 76V50', 'M48 76V40', 'M48 76V34']
const LEAVES = [
  [],
  ['M48 64c-9 0-13-5-13-11 8 0 13 4 13 11z', 'M48 64c9 0 13-5 13-11-8 0-13 4-13 11z'],
  [
    'M48 68c-11 0-17-5-18-13 10-1 17 4 18 13z',
    'M48 68c11 0 17-5 18-13-10-1-17 4-18 13z',
    'M48 54c-8-1-12-6-12-12 8 1 12 5 12 12z',
    'M48 54c8-1 12-6 12-12-8 1-12 5-12 12z'
  ],
  [
    'M48 70c-12 0-19-6-20-14 11-1 19 4 20 14z',
    'M48 70c12 0 19-6 20-14-11-1-19 4-20 14z',
    'M48 56c-9-1-14-6-14-13 9 1 14 6 14 13z',
    'M48 56c9-1 14-6 14-13-9 1-14 6-14 13z'
  ],
  [
    'M48 74c-15 0-24-7-25-18 14-1 24 6 25 18z',
    'M48 74c15 0 24-7 25-18-14-1-24 6-25 18z',
    'M48 60c-12-1-19-8-19-18 12 1 19 8 19 18z',
    'M48 60c12-1 19-8 19-18-12 1-19 8-19 18z',
    'M48 46c-7-3-10-9-9-16 7 3 10 8 9 16z',
    'M48 46c7-3 10-9 9-16-7 3-10 8-9 16z'
  ]
]
const BLOOMS: number[][][] = [[], [], [], [[48, 34, 6]], [[48, 26, 6], [34, 34, 4], [62, 34, 4]]]

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
.yc-growth-plant .soil { fill: var(--yc-fill); }
.yc-growth-plant .stem { fill: none; stroke: var(--yc-ink); stroke-width: 3; stroke-linecap: round; }
.yc-growth-plant .leaf { fill: var(--yc-lime); stroke: var(--yc-ink); stroke-width: 2; stroke-linejoin: round; }
.yc-growth-plant .bloom { fill: var(--yc-paper); stroke: var(--yc-ink); stroke-width: 2; }
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
