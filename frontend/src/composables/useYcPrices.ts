import { computed, ref } from 'vue'
import { getModelPlaza, type ModelPlazaGroup } from '@/api/modelPlaza'
import { useYcMoney } from '@/composables/useYcMoney'

const PER_MILLION = 1_000_000
const FEATURED = ['claude', 'gpt', 'gemini', 'grok', 'deepseek']

export interface YcPrice {
  name: string
  /** ¥ per million tokens actually charged. */
  input: number
  output: number
  /** Official price restated in ¥, null when unknown. */
  officialInput: number | null
  officialOutput: number | null
  /** Charged price as a share of the official one (0-1), null when unknown or not cheaper. */
  share: number | null
}

/** Final ¥ prices per model, with the multiplier and recharge ratio already applied. */
export function useYcPrices() {
  const { toYuan, cnyPerUsd } = useYcMoney()
  const groups = ref<ModelPlazaGroup[]>([])

  async function load() {
    try {
      groups.value = (await getModelPlaza()).groups
    } catch {
      groups.value = []
    }
  }

  const prices = computed<YcPrice[]>(() => {
    const best = new Map<string, YcPrice>()
    for (const g of groups.value) {
      const rate = g.user_rate_multiplier ?? g.rate_multiplier
      for (const m of g.models) {
        const p = m.pricing
        if (!p || p.input_price == null || p.output_price == null) continue
        const input = toYuan(p.input_price * rate * PER_MILLION)
        const known = best.get(m.name)
        if (known && known.input <= input) continue
        const o = m.official_pricing
        const officialInput = o?.input_price != null ? o.input_price * PER_MILLION * cnyPerUsd.value : null
        const officialOutput = o?.output_price != null ? o.output_price * PER_MILLION * cnyPerUsd.value : null
        const share = officialInput && input < officialInput ? input / officialInput : null
        best.set(m.name, {
          name: m.name,
          input,
          output: toYuan(p.output_price * rate * PER_MILLION),
          officialInput,
          officialOutput,
          share
        })
      }
    }
    const rank = (name: string) => {
      const i = FEATURED.findIndex((k) => name.toLowerCase().includes(k))
      return i === -1 ? FEATURED.length : i
    }
    return [...best.values()].sort((a, b) => rank(a.name) - rank(b.name) || a.name.localeCompare(b.name))
  })

  return { prices, load, cnyPerUsd }
}
