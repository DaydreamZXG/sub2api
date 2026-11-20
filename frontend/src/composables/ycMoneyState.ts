// Shared money state for the redesigned pages. Kept free of stores so any component can import it.
import { computed, ref } from 'vue'

// Balance units credited per ¥1 paid. Users see what they paid, so amounts are shown in ¥.
export const unitsPerYuan = ref(1)
// Reference rate used only to restate official USD prices in ¥ for comparison.
export const cnyPerUsd = ref(7.2)
// Pages that print stored amounts as they are may only call them yuan while the two are equal.
export const ycSign = computed(() => (unitsPerYuan.value === 1 ? '¥' : '$'))
