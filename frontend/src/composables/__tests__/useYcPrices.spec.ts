import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { ModelPlazaGroup } from '@/api/modelPlaza'
import { useYcPrices } from '../useYcPrices'

const { getModelPlaza } = vi.hoisted(() => ({ getModelPlaza: vi.fn() }))

vi.mock('@/api/modelPlaza', () => ({ getModelPlaza }))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ cachedPublicSettings: null }),
  useAuthStore: () => ({ isAuthenticated: false })
}))

// $1 per million input tokens and $2 per million output tokens before the line's multiplier.
function line(name: string, rate: number, extra: Partial<ModelPlazaGroup> = {}): ModelPlazaGroup {
  return {
    name,
    rate_multiplier: rate,
    subscription_type: 'standard',
    models: [{ name: 'claude-sonnet', pricing: { input_price: 1e-6, output_price: 2e-6 } }],
    ...extra
  } as ModelPlazaGroup
}

async function pricesFor(groups: ModelPlazaGroup[]) {
  getModelPlaza.mockResolvedValue({ groups })
  const { prices, load } = useYcPrices()
  await load()
  return prices.value
}

describe('useYcPrices with app-only lines', () => {
  beforeEach(() => getModelPlaza.mockReset())

  it('keeps the price anyone can get as the headline and quotes the app price beside it', async () => {
    const [model] = await pricesFor([line('app', 0.75, { app_only: true }), line('normal', 1)])

    expect([model.input, model.output]).toEqual([1, 2])
    expect(model.app).toEqual({ input: 0.75, output: 1.5, varies: false })
    expect(model.appOnly).toBe(false)
    expect(model.lines.map((l) => [l.name, l.app])).toEqual([['normal', false], ['app', true]])
  })

  it('flags an app price that can be higher for some requests', async () => {
    const [model] = await pricesFor([
      line('app', 0.75, { app_only: true, peak_rate_enabled: true, peak_rate_multiplier: 2 }),
      line('normal', 1)
    ])

    expect(model.varies).toBe(false)
    expect(model.app?.varies).toBe(true)
  })

  it('quotes no app price when the app line is not cheaper', async () => {
    const [model] = await pricesFor([line('app', 1.25, { app_only: true }), line('normal', 1)])

    expect(model.input).toBe(1)
    expect(model.app).toBeNull()
  })

  it('marks a model sold only through the app', async () => {
    const [model] = await pricesFor([line('app', 0.75, { app_only: true })])

    expect(model.input).toBe(0.75)
    expect(model.appOnly).toBe(true)
    expect(model.app).toBeNull()
  })

  it('does not call a model app-only when its app line also takes other tools through a fallback line', async () => {
    const [model] = await pricesFor([line('app', 0.75, { app_only: true, app_fallback: true })])

    expect(model.appOnly).toBe(false)
    expect(model.appHeadline).toBe(true)
  })

  it('quotes the app price when only the output is cheaper there', async () => {
    const [model] = await pricesFor([
      line('app', 1, {
        app_only: true,
        models: [{ name: 'claude-sonnet', pricing: { input_price: 1e-6, output_price: 1e-6 } }]
      } as Partial<ModelPlazaGroup>),
      line('normal', 1)
    ])

    expect(model.output).toBe(2)
    expect(model.app).toEqual({ input: 1, output: 1, varies: false })
  })

  it('does not call a model app-only when a monthly plan also sells it', async () => {
    const [model] = await pricesFor([line('app', 0.75, { app_only: true }), line('plan', 1, { subscription_type: 'subscription' })])

    expect(model.input).toBe(0.75)
    expect(model.appOnly).toBe(false)
    expect(model.appHeadline).toBe(true)
  })

  it('is unchanged when no line is app-only', async () => {
    const [model] = await pricesFor([line('cheap', 0.5), line('normal', 1)])

    expect(model.input).toBe(0.5)
    expect(model.app).toBeNull()
    expect(model.appOnly).toBe(false)
  })
})
