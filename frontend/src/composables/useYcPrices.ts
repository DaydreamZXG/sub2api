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
  /** Every line selling this model: open lines, then app-only, then plan-only, cheapest first. The headline price is the first one. */
  lines: YcPriceLine[]
  /** Cheapest app-only price, when it beats the headline; null when there is none. */
  app: { input: number; output: number; varies: boolean } | null
  /** True when every line selling the model is app-only, i.e. it is not sold outside the app. */
  appOnly: boolean
  /** True when the headline is an app-only price although a monthly plan also sells the model. */
  appHeadline: boolean
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
  /** Charged only to requests sent through the desktop app. */
  app: boolean
  /** An app line that also takes requests sent without the app, billed on another line. */
  appFallback: boolean
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
          app: g.app_only === true,
          appFallback: g.app_only === true && g.app_fallback === true,
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
          app: null,
          appOnly: false,
          appHeadline: false,
          usdInput: o?.input_price != null ? o.input_price * PER_MILLION : null,
          usdOutput: o?.output_price != null ? o.output_price * PER_MILLION : null
        })
      }
    }
    for (const price of best.values()) {
      // Open pay-as-you-go lines first: the headline must be a price anyone can get, with any tool.
      price.lines.sort((a, b) => Number(a.plan) - Number(b.plan) || Number(a.app) - Number(b.app) || a.input - b.input)
      const headline = price.lines[0]
      price.input = headline.input
      price.output = headline.output
      price.varies = headline.varies
      // A fallback line may be hidden from the list, yet it still lets other tools use the model.
      price.appOnly = price.lines.every((l) => l.app && !l.appFallback)
      price.appHeadline = headline.app && !price.appOnly
      const appLine = price.lines.find((l) => l.app && !l.plan)
      price.app =
        appLine && !headline.app && appLine.input < headline.input
          ? { input: appLine.input, output: appLine.output, varies: appLine.varies }
          : null
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
