import { officialUsdPerToken, usdText } from '@/utils/pricing'
import { describe, expect, it } from 'vitest'
import { formatScaled } from '../pricing'

describe('formatScaled', () => {
  it.each([
    [1e-10, 1, '$1e-10'],
    [1.25e-10, 1, '$1.25e-10'],
    [1e-20, 1, '$1e-20'],
    [1e-16, 1_000_000, '$1e-10'],
    [1e10, 1, '$10000000000'],
  ])('preserves the magnitude of %s scaled by %s', (value, scale, expected) => {
    expect(formatScaled(value, scale)).toBe(expected)
  })

  it.each([
    [0.000003, 1_000_000, 0, '$3'],
    [0.000003, 1_000_000, 2, '$3.00'],
    [1.25e-8, 1_000_000, 2, '$0.0125'],
    [5e-8, 1_000_000, 0, '$0.05'],
    [0, 1, 2, '$0.00'],
    [null, 1_000_000, 0, '-'],
  ])('retains ordinary pricing and padding for %s', (value, scale, digits, expected) => {
    expect(formatScaled(value, scale, digits)).toBe(expected)
  })
})

describe('usdText', () => {
  it('formats dollars with two decimals and drops conversion noise', () => {
    expect(usdText(0.15)).toBe('$0.15')
    expect(usdText(0.0015)).toBe('$0.0015')
    expect(usdText(0.15 * 6.6647)).toBe('$0.9997')
  })
})

describe('official prices', () => {
  const base = { input_price: 1.5e-7, output_price: 6e-7, cache_write_price: null, cache_read_price: null }

  it('uses the catalog dollars and ignores a yuan price', () => {
    expect(officialUsdPerToken({ ...base, cny_input_price: 1e-6 }, 'input_price')).toBe(1.5e-7)
    expect(officialUsdPerToken({ ...base, input_price: null, cny_input_price: 7e-6 }, 'input_price')).toBeNull()
    expect(officialUsdPerToken(base, 'cache_read_price')).toBeNull()
    expect(officialUsdPerToken(null, 'input_price')).toBeNull()
  })
})
