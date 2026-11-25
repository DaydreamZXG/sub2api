// Shared money state for the redesigned pages. Kept free of stores so any component can import it.
import { computed, ref } from 'vue'

// The site bills in US dollars. Balance units credited per $1 paid.
export const unitsPerDollar = ref(1)
export const ycSign = computed(() => '$')
