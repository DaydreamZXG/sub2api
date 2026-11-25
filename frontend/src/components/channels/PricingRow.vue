<template>
  <div class="flex justify-between gap-2">
    <span class="text-gray-500 dark:text-gray-400">{{ label }}</span>
    <span class="font-mono">{{ display }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { unitsPerDollar } from '@/composables/ycMoneyState'
import { usdText } from '@/utils/pricing'

const props = withDefaults(
  defineProps<{
    label: string
    value: number | null
    unit: string
    scale: number
  }>(),
  { value: null }
)

const display = computed(() =>
  props.value == null
    ? '-'
    : `${usdText((props.value * props.scale) / unitsPerDollar.value)} ${props.unit}`
)
</script>
