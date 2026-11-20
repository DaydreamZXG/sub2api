type Translate = (key: string, named: Record<string, unknown>) => string

/**
 * A price multiplier said the way people talk about prices: "约 8 折" / "贵 20%".
 * Empty for the standard price (1) and for values that are not a positive number.
 */
export function priceChangeLabel(t: Translate, multiplier: number | null | undefined): string {
  if (multiplier == null || !(multiplier > 0) || multiplier === 1) return ''
  const pct = Math.round(Math.abs(1 - multiplier) * 100)
  return multiplier < 1
    ? t('yc.priceWords.cheaper', { zhe: parseFloat((multiplier * 10).toFixed(1)), pct })
    : t('yc.priceWords.dearer', { pct })
}
