import { describe, expect, it } from 'vitest'
import { priceChangeLabel } from '../ycPriceWords'

const t = (key: string, named: Record<string, unknown>) => `${key} ${JSON.stringify(named)}`

describe('priceChangeLabel', () => {
  it.each([
    [0.8, 'yc.priceWords.cheaper {"zhe":8,"pct":20}'],
    [0.85, 'yc.priceWords.cheaper {"zhe":8.5,"pct":15}'],
    [1.2, 'yc.priceWords.dearer {"pct":20}'],
    [2, 'yc.priceWords.dearer {"pct":100}'],
  ])('says a multiplier of %s as a price change', (multiplier, expected) => {
    expect(priceChangeLabel(t, multiplier)).toBe(expected)
  })

  it.each([1, 0, -1, null, undefined, Number.NaN])('says nothing for %s', (multiplier) => {
    expect(priceChangeLabel(t, multiplier)).toBe('')
  })
})
