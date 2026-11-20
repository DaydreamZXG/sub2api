type Translate = (key: string, named: Record<string, unknown>) => string

/**
 * A price multiplier said the way people talk about prices: "约 8 折" / "贵 20%" / "免费" for 0.
 * Empty for the standard price (1) and for values that are not a number from 0 up.
 */
export function priceChangeLabel(t: Translate, multiplier: number | null | undefined): string {
  if (multiplier === 0) return t('yc.priceWords.free', {})
  if (multiplier == null || !(multiplier > 0) || multiplier === 1) return ''
  const pct = Math.round(Math.abs(1 - multiplier) * 100)
  return multiplier < 1
    ? t('yc.priceWords.cheaper', { zhe: parseFloat((multiplier * 10).toFixed(1)), pct })
    : t('yc.priceWords.dearer', { pct })
}
