<template>
  <div class="pzt" :style="{ '--pzt-ref-label': JSON.stringify(t('modelPlaza.table.officialShort')) }">
    <table class="pzt-table">
      <thead>
        <tr>
          <th rowspan="2" class="pzt-th-model">{{ t('modelPlaza.table.model') }}</th>
          <th colspan="3" class="pzt-th-group pzt-paid">
            {{ t('modelPlaza.table.paidPrice') }}
            <span class="pzt-unit">{{ t('modelPlaza.table.unitPerMillion') }}</span>
          </th>
          <th colspan="3" class="pzt-th-group pzt-ref">
            {{ t('modelPlaza.table.officialPrice') }}
            <span class="pzt-unit">{{ t('modelPlaza.table.unitPerMillion') }}</span>
          </th>
          <th rowspan="2" class="pzt-th-off">{{ t('modelPlaza.table.rate') }}</th>
        </tr>
        <tr class="pzt-sub">
          <th class="pzt-paid">{{ t('modelPlaza.table.input') }}</th>
          <th class="pzt-paid">{{ t('modelPlaza.table.output') }}</th>
          <th class="pzt-paid">{{ t('modelPlaza.table.cache') }}</th>
          <th class="pzt-ref">{{ t('modelPlaza.table.input') }}</th>
          <th class="pzt-ref">{{ t('modelPlaza.table.output') }}</th>
          <th class="pzt-ref">{{ t('modelPlaza.table.cache') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="{ model: m, period, key } in rows" :key="key" :class="{ 'pzt-period': period }">
          <td class="pzt-model">
            <div class="pzt-model-in">
              <span class="pzt-name">{{ m.name }}</span>
              <span v-if="period" class="pzt-chip pzt-chip-mono" :title="timePricingRowHint(m)">
                <span v-if="m.time_pricing?.weekdays_only" class="pzt-chip-word">{{
                  t('modelPlaza.table.timePricingWeekdays')
                }}</span>
                {{ formatTimeWindow(period) }}
              </span>
              <span v-if="platform && m.platform !== platform" class="pzt-chip">{{ platformLabel(m.platform) }}</span>
              <span v-if="billingMode(m) !== BILLING_MODE_TOKEN" class="pzt-chip">{{ billingModeLabel(m) }}</span>
              <span
                v-if="m.long_context_basis === 'marginal'"
                class="pzt-chip"
                :title="t('modelPlaza.table.tierHintMarginal')"
              >
                {{ t('modelPlaza.table.marginalBadge') }}
              </span>
              <span
                v-for="([effort, multiplier]) in reasoningEffortMultipliers(m)"
                :key="effort"
                class="pzt-chip pzt-chip-warn"
                :title="t('modelPlaza.table.reasoningMultiplierHint', { effort, multiplier })"
                :data-reasoning-effort="effort"
              >
                {{ t('modelPlaza.table.reasoningMultiplierBadge', { effort, change: priceChangeLabel(t, multiplier) || t('modelPlaza.table.standardPrice') }) }}
              </span>
            </div>
          </td>

          <!-- Charged per token: input / output / cache, one line per context tier. -->
          <template v-if="billingMode(m) === BILLING_MODE_TOKEN">
            <td class="pzt-paid pzt-num" :data-label="t('modelPlaza.table.input')">
              <template v-if="tokenIntervals(m).length">
                <div v-for="(iv, idx) in tokenIntervals(m)" :key="idx" class="pzt-tier">
                  <span class="pzt-dim" :title="tierHint(m)">{{ tierLabel(iv) }}</span>
                  {{ paidPerMillion(iv.input_price, period) }}
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.input_price, period) }}</template>
            </td>
            <td class="pzt-paid pzt-num" :data-label="t('modelPlaza.table.output')">
              <template v-if="tokenIntervals(m).length">
                <div v-for="(iv, idx) in tokenIntervals(m)" :key="idx" class="pzt-tier" :title="tierHint(m)">
                  {{ paidPerMillion(iv.output_price, period) }}
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.output_price, period) }}</template>
            </td>
            <td class="pzt-paid pzt-cache" :data-label="t('modelPlaza.table.cache')">
              <template v-if="hasTierCachePricing(tokenIntervals(m))">
                <div v-for="(iv, idx) in tokenIntervals(m)" :key="idx" class="pzt-tier" :title="tierHint(m)">
                <template v-if="iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_write_12h_price != null || iv.cache_read_price != null">
                  <span class="pzt-dim">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                  {{ paidPerMillion(iv.cache_write_price, period) }}
                  <template v-if="iv.cache_write_1h_price != null"
                    ><span class="pzt-dim"> (1h </span>{{ paidPerMillion(iv.cache_write_1h_price, period)
                    }}<span class="pzt-dim">)</span></template
                  ><template v-if="iv.cache_write_12h_price != null"
                    ><span class="pzt-dim"> (12h </span>{{ paidPerMillion(iv.cache_write_12h_price, period)
                    }}<span class="pzt-dim">)</span></template
                  >
                  <span class="pzt-dim">{{ t('modelPlaza.table.cacheReadShort') }}</span>
                  {{ paidPerMillion(iv.cache_read_price, period) }}
                </template>
                <span v-else class="pzt-dim">-</span>
                </div>
              </template>
              <div v-else-if="m.pricing && hasCachePricing(m)" class="pzt-flat">
              <div>
                <span class="pzt-dim">{{ t('modelPlaza.table.cacheWrite') }}</span>
                {{ paidPerMillion(m.pricing.cache_write_price, period)
                }}<template v-if="m.pricing.cache_write_1h_price != null"
                  ><span class="pzt-dim"> (1h </span>{{ paidPerMillion(m.pricing.cache_write_1h_price, period)
                  }}<span class="pzt-dim">)</span></template
                ><template v-if="m.pricing.cache_write_12h_price != null"
                  ><span class="pzt-dim"> (12h </span>{{ paidPerMillion(m.pricing.cache_write_12h_price, period)
                  }}<span class="pzt-dim">)</span></template
                >
              </div>
              <div>
                <span class="pzt-dim">{{ t('modelPlaza.table.cacheRead') }}</span>
                {{ paidPerMillion(m.pricing.cache_read_price, period) }}
              </div>
              </div>
              <span v-else class="pzt-dim">-</span>
            </td>
          </template>

          <!-- Charged per request or per image: one merged cell. -->
          <template v-else>
            <td colspan="3" class="pzt-paid pzt-num pzt-request" :data-label="billingModeLabel(m)">
              <div v-if="requestIntervals(m).length" class="pzt-request-tiers">
                <span v-for="(iv, idx) in requestIntervals(m)" :key="idx" class="pzt-request-tier">
                  <span class="pzt-dim">{{ tierLabel(iv) }}</span>
                  {{ paidRequestPrice(m, iv.per_request_price)
                  }}<span class="pzt-dim">{{ perUnitSuffix(m) }}</span>
                </span>
              </div>
              <template v-else-if="m.pricing?.per_request_price != null">
                {{ paidRequestPrice(m, m.pricing.per_request_price) }}
                <span class="pzt-dim">{{ perUnitSuffix(m) }}</span>
              </template>
              <span v-else class="pzt-dim">-</span>
            </td>
          </template>

          <!-- Official list price, for comparison only. -->
          <td class="pzt-ref pzt-num" :data-label="t('modelPlaza.table.input')">
            <template v-if="officialIntervals(m).length">
              <div v-for="(iv, idx) in officialIntervals(m)" :key="idx" class="pzt-tier">
                <span class="pzt-dim" :title="t('modelPlaza.table.tierHint')">{{ tierLabel(iv) }}</span>
                {{ official(iv.input_price) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.input_price) }}</template>
          </td>
          <td class="pzt-ref pzt-num" :data-label="t('modelPlaza.table.output')">
            <template v-if="officialIntervals(m).length">
              <div v-for="(iv, idx) in officialIntervals(m)" :key="idx" class="pzt-tier" :title="t('modelPlaza.table.tierHint')">
                {{ official(iv.output_price) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.output_price) }}</template>
          </td>
          <td class="pzt-ref pzt-cache" :data-label="t('modelPlaza.table.cache')">
            <template v-if="hasTierCachePricing(officialIntervals(m))">
              <div v-for="(iv, idx) in officialIntervals(m)" :key="idx" class="pzt-tier" :title="t('modelPlaza.table.tierHint')">
                <template v-if="iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_write_12h_price != null || iv.cache_read_price != null">
                  <span class="pzt-dim">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                  {{ official(iv.cache_write_price) }}
                  <template v-if="iv.cache_write_1h_price != null"
                    ><span class="pzt-dim"> (1h </span>{{ official(iv.cache_write_1h_price)
                    }}<span class="pzt-dim">)</span></template
                  ><template v-if="iv.cache_write_12h_price != null"
                    ><span class="pzt-dim"> (12h </span>{{ official(iv.cache_write_12h_price)
                    }}<span class="pzt-dim">)</span></template
                  >
                  <span class="pzt-dim">{{ t('modelPlaza.table.cacheReadShort') }}</span>
                  {{ official(iv.cache_read_price) }}
                </template>
                <span v-else class="pzt-dim">-</span>
              </div>
            </template>
            <div v-else-if="m.official_pricing && hasOfficialCache(m.official_pricing)" class="pzt-flat">
              <div>
                <span class="pzt-dim">{{ t('modelPlaza.table.cacheWrite') }}</span>
                {{ official(m.official_pricing.cache_write_price)
                }}<template v-if="m.official_pricing.cache_write_1h_price != null"
                  ><span class="pzt-dim"> (1h </span>{{ official(m.official_pricing.cache_write_1h_price)
                  }}<span class="pzt-dim">)</span></template
                ><template v-if="m.official_pricing.cache_write_12h_price != null"
                  ><span class="pzt-dim"> (12h </span>{{ official(m.official_pricing.cache_write_12h_price)
                  }}<span class="pzt-dim">)</span></template
                >
              </div>
              <div>
                <span class="pzt-dim">{{ t('modelPlaza.table.cacheRead') }}</span>
                {{ official(m.official_pricing.cache_read_price) }}
              </div>
            </div>
            <span v-else class="pzt-dim">-</span>
          </td>

          <!-- The price next to the standard one, in words; the raw multiplier is in the tooltip. -->
          <td class="pzt-off">
            <span
              v-if="period"
              class="pzt-off-tag"
              :class="{ 'is-plain': !priceChangeLabel(t, period.multiplier) }"
              :data-rate="periodRate(period)"
              :title="t('modelPlaza.table.timePricingRateHint', { rate: effectiveRate, multiplier: period.multiplier })"
              >{{ periodLabel(period) }}</span
            >
            <template v-else-if="!usesIndependentMediaRate(m) && hasCustomRate">
              <s class="pzt-off-was" :data-rate-was="rateMultiplier">{{ rateLabel(rateMultiplier) }}</s>
              <span class="pzt-off-tag" :data-rate="effectiveRate" :title="t('modelPlaza.table.rawRate', { rate: effectiveRate })">{{ rateLabel(effectiveRate) }}</span>
            </template>
            <span
              v-else
              class="pzt-off-tag"
              :class="{ 'is-plain': !priceChangeLabel(t, rowRate(m)) }"
              :data-rate="rowRate(m)"
              :title="t('modelPlaza.table.rawRate', { rate: rowRate(m) })"
              >{{ rateLabel(rowRate(m)) }}</span
            >
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { resolveIntervalPrices, yuanText } from '@/utils/pricing'
import { platformLabel } from '@/utils/platformColors'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import { cnyPerUsd, unitsPerYuan } from '@/composables/ycMoneyState'
import {
  BILLING_MODE_TOKEN,
  BILLING_MODE_IMAGE,
  BILLING_MODE_VIDEO,
  REASONING_EFFORT_LEVELS,
  type BillingMode
} from '@/constants/channel'
import type { PlazaModel, PlazaTimePricingPeriod } from '@/api/modelPlaza'
import type { UserPricingInterval } from '@/api/channels'

function reasoningEffortMultipliers(model: PlazaModel): [string, number][] {
  const multipliers = model.pricing?.reasoning_effort_multipliers
  return REASONING_EFFORT_LEVELS.flatMap(effort => {
    const multiplier = multipliers?.[effort]
    return typeof multiplier === 'number' && Number.isFinite(multiplier) && multiplier > 0
      ? [[effort, multiplier] as [string, number]]
      : []
  })
}

const props = defineProps<{
  models: PlazaModel[]
  /** 分组平台;实付分区底色随平台着色,未知平台回退品牌青。 */
  platform?: string
  /** 分组默认倍率。 */
  rateMultiplier: number
  /** 用户专属倍率;与默认不同,实付价按此计算并划线展示原倍率。 */
  userRateMultiplier?: number | null
  /** 生图独立倍率:true 时图片计费模型的实付倍率取 imageRateMultiplier,不取分组/专属倍率。 */
  imageRateIndependent?: boolean
  imageRateMultiplier?: number | null
  /** 视频独立倍率开启时，覆盖分组/用户专属倍率，与后端视频计费保持一致。 */
  videoRateIndependent?: boolean
  videoRateMultiplier?: number | null
  /**
   * 高峰窗口描述(含倍率与服务器时区标注),空串/缺省 = 分组未启用高峰。
   * 表格所有价格均为不含高峰因子的口径,该窗口仅用于分时时段行的 tooltip 披露:
   * 与高峰重叠的部分实付还会再乘高峰倍率。
   */
  peakWindow?: string
  peakRateMultiplier?: number | null
}>()

const { t } = useI18n()

const PER_MILLION = 1_000_000

/**
 * 展示顺序:
 * 1. token 计费的排在前,按图/按次计费的沉到末尾——它们的官方 token 价与实付的按张/按次价不同量纲,混排无意义;
 * 2. 组内按官方输出价从高到低,无官方价的排最后;
 * 3. 同价按名称降序(新版本号在前,如 gpt-5.6 先于 gpt-5.5)。
 */
const sortedModels = computed(() => {
  return [...props.models].sort((a, b) => {
    const ta = billingMode(a) === BILLING_MODE_TOKEN
    const tb = billingMode(b) === BILLING_MODE_TOKEN
    if (ta !== tb) return ta ? -1 : 1
    const pa = a.official_pricing?.output_price ?? null
    const pb = b.official_pricing?.output_price ?? null
    if (pa != null && pb != null && pa !== pb) return pb - pa
    if (pa != null && pb == null) return -1
    if (pa == null && pb != null) return 1
    return b.name.localeCompare(a.name)
  })
})

const effectiveRate = computed(() => props.userRateMultiplier ?? props.rateMultiplier)
const hasCustomRate = computed(
  () => props.userRateMultiplier != null && props.userRateMultiplier !== props.rateMultiplier
)

function billingMode(m: PlazaModel): BillingMode {
  return (m.pricing?.billing_mode || BILLING_MODE_TOKEN) as BillingMode
}

function billingModeLabel(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perImage')
    : t('modelPlaza.table.perRequest')
}

/** 表格行:每个模型一行标准价;配置了分时倍率的模型再按时段各加一行。 */
interface PlazaRow {
  model: PlazaModel
  period: PlazaTimePricingPeriod | null
  key: string
}

const rows = computed<PlazaRow[]>(() =>
  sortedModels.value.flatMap((m) => {
    const base: PlazaRow = { model: m, period: null, key: `${m.platform}:${m.name}` }
    const periodRows = timePeriods(m).map<PlazaRow>((p, idx) => ({
      model: m,
      period: p,
      key: `${m.platform}:${m.name}:${idx}`
    }))
    return [base, ...periodRows]
  })
)

/** 时段行的生效倍率 = 生效倍率 × 时段倍率(去掉浮点噪声)。 */
function periodRate(period: PlazaTimePricingPeriod): number {
  return Math.round(effectiveRate.value * period.multiplier * 1000) / 1000
}

/** 实付价 = 渠道单价 × 生效倍率(时段行再乘时段倍率),按 $/1M token 展示。 */
function paidPerMillion(value: number | null | undefined, period: PlazaTimePricingPeriod | null = null): string {
  if (value == null) return '-'
  const rate = period ? periodRate(period) : effectiveRate.value
  return yuanText((value * rate * PER_MILLION) / unitsPerYuan.value)
}

/** 图片和视频分别使用对应的独立倍率开关。 */
function usesIndependentMediaRate(m: PlazaModel): boolean {
  return (billingMode(m) === BILLING_MODE_IMAGE && props.imageRateIndependent === true) ||
    (billingMode(m) === BILLING_MODE_VIDEO && props.videoRateIndependent === true)
}

/** 非 token 行的生效倍率，图片和视频可分别覆盖通用倍率。 */
function requestRate(m: PlazaModel): number {
  if (billingMode(m) === BILLING_MODE_VIDEO && props.videoRateIndependent === true) {
    return Math.max(0, props.videoRateMultiplier ?? 1)
  }
  return billingMode(m) === BILLING_MODE_IMAGE && props.imageRateIndependent === true
    ? (props.imageRateMultiplier ?? 1)
    : effectiveRate.value
}

/** 非 token 单价乘该行生效倍率，不换算为每百万 token。 */
function paidRequestPrice(m: PlazaModel, value: number | null | undefined): string {
  if (value == null) return '-'
  return yuanText((value * requestRate(m)) / unitsPerYuan.value)
}

/** 官方参考价不乘倍率。 */
function official(value: number | null | undefined): string {
  if (value == null) return '-'
  return yuanText(value * PER_MILLION * cnyPerUsd.value)
}

/** A multiplier in words; the standard price has its own label. */
function rateLabel(rate: number): string {
  return priceChangeLabel(t, rate) || t('modelPlaza.table.standardPrice')
}

/** The multiplier a standard (non time-of-day) row is charged at. */
function rowRate(m: PlazaModel): number {
  return requestRate(m)
}

/** How a time-of-day row compares with the model's standard row. */
function periodLabel(period: PlazaTimePricingPeriod): string {
  const change = priceChangeLabel(t, period.multiplier)
  return change ? t('modelPlaza.table.periodChange', { change }) : t('modelPlaza.table.standardPrice')
}

/** 非 token 计费的单位后缀:按图片 → “/ 张”,按次 → “/ 次”。 */
function perUnitSuffix(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perUnitImage')
    : t('modelPlaza.table.perUnitRequest')
}

function hasCachePricing(m: PlazaModel): boolean {
  return m.pricing?.cache_write_price != null || m.pricing?.cache_write_1h_price != null || m.pricing?.cache_write_12h_price != null || m.pricing?.cache_read_price != null
}

function hasOfficialCache(o: NonNullable<PlazaModel['official_pricing']>): boolean {
  return o.cache_write_price != null || o.cache_read_price != null || o.cache_write_1h_price != null || o.cache_write_12h_price != null
}

/** 分时倍率时段(后端只给出倍率 ≠ 1 的时段,已升序)。 */
function timePeriods(m: PlazaModel): PlazaTimePricingPeriod[] {
  return m.time_pricing?.periods ?? []
}

/**
 * 时段行 tooltip:仅工作日生效的配置换用带周末回落说明的文案;
 * 分组启用高峰倍率时追加披露——本行价格不含高峰因子,与高峰窗口重叠的部分实付再乘高峰倍率。
 */
function timePricingRowHint(m: PlazaModel): string {
  const key = m.time_pricing?.weekdays_only
    ? 'modelPlaza.table.timePricingRowHintWeekdays'
    : 'modelPlaza.table.timePricingRowHint'
  let hint = t(key, { timezone: m.time_pricing?.timezone })
  if (props.peakWindow) {
    hint += t('modelPlaza.table.timePricingRowHintPeak', {
      window: props.peakWindow,
      change: rateLabel(props.peakRateMultiplier ?? 1)
    })
  }
  return hint
}

/** “00:30–08:30”;整分钟的 HH:mm:ss 省略秒。 */
function formatTimeWindow(p: PlazaTimePricingPeriod): string {
  const clock = (v: string) => v.replace(/^(\d{2}:\d{2}):00$/, '$1')
  return `${clock(p.start_time)}–${clock(p.end_time)}`
}

/** 上下文档位按下限升序展示(后端已升序,此处兜底)。 */
function sortByContext(intervals: UserPricingInterval[]): UserPricingInterval[] {
  return [...intervals].sort((a, b) => a.min_tokens - b.min_tokens)
}

/** token 模式的阶梯定价(内联进输入/输出/缓存列)。 */
function tokenIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.pricing?.intervals ?? []).map(iv => resolveIntervalPrices(iv, m.pricing!))
}

/** 官方阶梯(后端按目录规则合成,不受分组开关影响)。 */
function officialIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.official_pricing?.intervals ?? [])
}

/** 任一档带缓存价才按档渲染缓存列;否则沿用平价的写入/读取两行。 */
function hasTierCachePricing(intervals: UserPricingInterval[]): boolean {
  return intervals.some((iv) =>
    iv.cache_write_price != null || iv.cache_write_1h_price != null || iv.cache_write_12h_price != null || iv.cache_read_price != null ||
    iv.cache_write_multiplier != null || iv.cache_read_multiplier != null
  )
}

/** 档位说明:整单按档计价,或(平台旧规则)仅超出部分按档计价。 */
function tierHint(m: PlazaModel): string {
  return m.long_context_basis === 'marginal'
    ? t('modelPlaza.table.tierHintMarginal')
    : t('modelPlaza.table.tierHint')
}


/** 按次/按图模式的阶梯定价(仅保留配了按次价的档位)。 */
function requestIntervals(m: PlazaModel): UserPricingInterval[] {
  return (m.pricing?.intervals ?? []).filter((iv) => iv.per_request_price != null)
}

/**
 * 档位标签:优先后端/管理员给出的 tier_label,否则按区间生成统一形态——
 * 有上限为「≤上限」,末档为「>下限」;档位升序排列,相邻的 ≤100K / ≤200K 即表示 (100K,200K]。
 */
function tierLabel(iv: UserPricingInterval): string {
  if (iv.tier_label) return iv.tier_label
  const { min_tokens: min, max_tokens: max } = iv
  return max == null ? `>${formatTokenCount(min)}` : `≤${formatTokenCount(max)}`
}

function formatTokenCount(n: number): string {
  if (n >= 1_000_000) return `${trimZero(n / 1_000_000)}M`
  if (n >= 1_000) return `${trimZero(n / 1_000)}K`
  return String(n)
}

function trimZero(n: number): string {
  return String(Math.round(n * 100) / 100)
}
</script>

<style scoped>
.pzt { border-top: 1px solid var(--yc-line); overflow-x: auto; }
.pzt-table { width: 100%; border-collapse: collapse; font-size: 14px; font-variant-numeric: tabular-nums; color: var(--yc-ink); }
.pzt-table th { padding: 10px 12px 4px; font-size: 12.5px; font-weight: 650; text-align: left; color: var(--yc-mut); white-space: nowrap; }
.pzt-sub th { padding: 2px 12px 10px; font-weight: 500; }
.pzt-th-model, .pzt-model { padding-left: 24px !important; }
.pzt-th-off, .pzt-off { padding-right: 24px !important; text-align: right !important; }
.pzt-th-group { color: var(--yc-ink) !important; }
.pzt-unit { margin-left: 6px; font-weight: 400; color: var(--yc-dim); }
.pzt-table td { padding: 13px 12px; border-top: 1px solid var(--yc-line); vertical-align: middle; }
.pzt-table tbody tr { transition: background-color 0.15s ease; }
.pzt-table tbody tr:hover { background: var(--yc-fill); }
.pzt-period td { border-top-style: dashed; }
.pzt-period .pzt-name { color: var(--yc-mut); font-weight: 500; }
.pzt-model { width: 26%; }
.pzt-model-in { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.pzt-name { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13.5px; font-weight: 600; overflow-wrap: anywhere; }
.pzt-chip { display: inline-flex; align-items: center; height: 20px; padding: 0 7px; border-radius: 6px; background: var(--yc-fill); font-size: 11.5px; font-weight: 500; color: var(--yc-mut); white-space: nowrap; }
.pzt-chip-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.pzt-chip-word { margin-right: 4px; font-family: inherit; }
.pzt-chip-warn { border: 1px solid var(--yc-line); background: transparent; }
.pzt-paid { background: color-mix(in srgb, var(--yc-lime) 9%, transparent); }
.pzt-num { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-weight: 650; white-space: nowrap; }
.pzt-ref { color: var(--yc-mut); }
.pzt-ref.pzt-num { font-weight: 400; }
.pzt-cache { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12.5px; line-height: 1.55; }
.pzt-tier, .pzt-flat { font-size: 12.5px; line-height: 1.6; white-space: nowrap; }
.pzt-dim { font-family: inherit; font-weight: 400; color: var(--yc-dim); }
.pzt-request-tiers { display: flex; flex-wrap: wrap; gap: 4px 14px; font-size: 12.5px; }
.pzt-off { white-space: nowrap; }
.pzt-off-tag { display: inline-flex; align-items: center; height: 24px; padding: 0 10px; border-radius: 999px; background: var(--yc-lime); font-size: 12.5px; font-weight: 650; color: var(--yc-on-lime); cursor: default; }
.pzt-off-tag.is-plain { background: transparent; padding: 0; font-weight: 500; color: var(--yc-dim); }
.pzt-off-was { margin-right: 6px; font-size: 12.5px; color: var(--yc-dim); }

/* Narrow screens: each row becomes a small card, so nothing scrolls sideways. */
@media (max-width: 860px) {
  .pzt { overflow-x: visible; }
  .pzt-table thead { display: none; }
  .pzt-table, .pzt-table tbody { display: block; }
  .pzt-table tbody tr { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px 10px; padding: 14px 16px; border-top: 1px solid var(--yc-line); }
  .pzt-table tbody tr:first-child { border-top: 0; }
  .pzt-table tbody tr.pzt-period { border-top-style: dashed; }
  .pzt-table td { padding: 0 !important; border: 0; background: transparent; text-align: left !important; white-space: normal; }
  .pzt-model { grid-column: 1 / 3; width: auto; }
  .pzt-off { grid-column: 3; grid-row: 1; justify-self: end; }
  .pzt-request { grid-column: 1 / -1; }
  .pzt-table td[data-label]::before { content: attr(data-label); display: block; margin-bottom: 2px; font-family: ui-sans-serif, system-ui, sans-serif; font-size: 11.5px; font-weight: 500; color: var(--yc-dim); }
  .pzt-ref { font-size: 12.5px; }
  .pzt-table td.pzt-ref[data-label]::before { content: var(--pzt-ref-label) " " attr(data-label); }
  .pzt-tier, .pzt-flat { white-space: normal; }
}
@media (prefers-reduced-motion: reduce) { .pzt-table tbody tr { transition: none; } }
</style>
