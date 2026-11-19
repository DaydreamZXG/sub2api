<template>
  <svg class="yc-plant" viewBox="10 14 76 76" role="img" :aria-label="label">
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
</template>

<script setup lang="ts">
// A plant drawn in five stages, 0 (seed) to 4 (full grown).
defineProps<{ stage: number; label: string }>()

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
</script>

<style scoped>
.yc-plant .soil { fill: var(--yc-fill); }
.yc-plant .stem { fill: none; stroke: var(--yc-ink); stroke-width: 3; stroke-linecap: round; }
.yc-plant .leaf { fill: var(--yc-lime); stroke: var(--yc-ink); stroke-width: 2; stroke-linejoin: round; }
.yc-plant .bloom { fill: var(--yc-paper); stroke: var(--yc-ink); stroke-width: 2; }
</style>
