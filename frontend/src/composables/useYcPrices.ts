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
  /** Every line selling this model: pay-as-you-go before plan-only, cheapest first. The headline price is the first one. */
  lines: YcPriceLine[]
  /** Official list price in USD per million tokens, for those who want the raw numbers. */
  usdInput: number | null
  usdOutput: number | null
  /** True when the headline line charges more than the quoted price for some requests. */
  varies: boolean
}

export interface YcPriceLine {
  name: string
  input: number
  output: number
  cacheRead: number | null
  multiplier: number
  /** Sold through a monthly plan, so only subscribers can use it. */
  plan: boolean
  /** The price is the standard tier; long context, time of day or reasoning effort can raise it. */
  varies: boolean
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
        const line: YcPriceLine = {
          name: g.name,
          input: toYuan(p.input_price * rate * PER_MILLION),
          output: toYuan(p.output_price * rate * PER_MILLION),
          cacheRead: p.cache_read_price != null ? toYuan(p.cache_read_price * rate * PER_MILLION) : null,
          multiplier: rate,
          plan: g.subscription_type === 'subscription',
          varies:
            (g.long_context_pricing_enabled && (p.intervals?.length ?? 0) > 1) ||
            (m.time_pricing?.periods ?? []).some((period) => period.multiplier > 1) ||
            (g.peak_rate_enabled && g.peak_rate_multiplier > 1) ||
            Object.values(p.reasoning_effort_multipliers ?? {}).some((x) => x > 1)
        }
        const known = best.get(m.name)
        if (known) {
          known.lines.push(line)
          continue
        }
        const o = m.official_pricing
        best.set(m.name, {
          name: m.name,
          input: line.input,
          output: line.output,
          officialInput: o?.input_price != null ? o.input_price * PER_MILLION * cnyPerUsd.value : null,
          officialOutput: o?.output_price != null ? o.output_price * PER_MILLION * cnyPerUsd.value : null,
          share: null,
          varies: false,
          lines: [line],
          usdInput: o?.input_price != null ? o.input_price * PER_MILLION : null,
          usdOutput: o?.output_price != null ? o.output_price * PER_MILLION : null
        })
      }
    }
    for (const price of best.values()) {
      // Pay-as-you-go lines first: the headline must be a price anyone can get.
      price.lines.sort((a, b) => Number(a.plan) - Number(b.plan) || a.input - b.input)
      price.input = price.lines[0].input
      price.output = price.lines[0].output
      price.varies = price.lines[0].varies
      price.share = price.officialInput && price.input < price.officialInput ? price.input / price.officialInput : null
    }
    const rank = (name: string) => {
      const i = FEATURED.findIndex((k) => name.toLowerCase().includes(k))
      return i === -1 ? FEATURED.length : i
    }
    return [...best.values()].sort((a, b) => rank(a.name) - rank(b.name) || a.name.localeCompare(b.name))
  })

  return { prices, load, cnyPerUsd }
}
