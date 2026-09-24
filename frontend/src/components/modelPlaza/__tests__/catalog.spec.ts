import { describe, expect, it } from 'vitest'
import {
  buildCatalog,
  filterCatalog,
  minimumTokenPrice,
  modelProvider,
  tokenPrice
} from '../catalog'
import type { ModelPlazaGroup, PlazaModel } from '@/api/modelPlaza'

const model = (name = 'gpt-example', platform = 'openai'): PlazaModel => ({
  name,
  platform,
  pricing: null,
  official_pricing: null
})
const group = (id: number, models: PlazaModel[]): ModelPlazaGroup =>
  ({
    id,
    name: `Group ${id}`,
    platform: 'openai',
    rate_multiplier: 2,
    models
  }) as ModelPlazaGroup

describe('model plaza catalog', () => {
  it('combines models across groups while preserving different protocols', () => {
    const result = buildCatalog([
      group(1, [model(), model()]),
      group(2, [model(), model('gpt-example', 'anthropic')])
    ])
    expect(result).toHaveLength(2)
    expect(result[0].routes).toHaveLength(2)
  })
  it('distinguishes unknown, non-token and zero prices', () => {
    const m = model()
    expect(
      tokenPrice({ group: group(1, [m]), model: m }, 'input_price')
    ).toBeNull()
    m.pricing = {
      billing_mode: 'token',
      input_price: 0,
      output_price: 0.000002
    } as NonNullable<PlazaModel['pricing']>
    expect(tokenPrice({ group: group(1, [m]), model: m }, 'input_price')).toBe(
      0
    )
    m.pricing.billing_mode = 'image'
    expect(
      tokenPrice({ group: group(1, [m]), model: m }, 'output_price')
    ).toBeNull()
  })
  it('applies personal rates once and converts per-token units', () => {
    const m = model()
    m.pricing = { billing_mode: 'token', input_price: 0.000002 } as NonNullable<
      PlazaModel['pricing']
    >
    const g = group(1, [m])
    g.user_rate_multiplier = 0.5
    expect(minimumTokenPrice(buildCatalog([g])[0], 'input_price')).toBe(1)
  })
  it('searches model and group while applying provider filters', () => {
    const list = buildCatalog([
      group(1, [model('claude-example', 'openai'), model('gpt-example')])
    ])
    expect(filterCatalog(list, 'anthropic', 'CLAUDE group 1')).toHaveLength(1)
    expect(filterCatalog(list, 'openai', 'claude')).toHaveLength(0)
    expect(filterCatalog(list, 'all', 'missing')).toHaveLength(0)
  })
  it('calculates cache quotes with personal rates and preserves zero versus unavailable prices', () => {
    const m = model()
    m.pricing = { billing_mode: 'token', cache_read_price: 2e-8, cache_write_price: 0,
      cache_write_1h_price: 2e-6 } as NonNullable<PlazaModel['pricing']>
    const g = group(1, [m])
    g.user_rate_multiplier = 0.4
    const item = buildCatalog([g])[0]
    expect(minimumTokenPrice(item, 'cache_read_price')).toBeCloseTo(0.008)
    expect(minimumTokenPrice(item, 'cache_write_price')).toBe(0)
    expect(minimumTokenPrice(item, 'cache_write_1h_price')).toBeCloseTo(0.8)
    m.pricing.cache_write_price = null
    expect(minimumTokenPrice(item, 'cache_write_price')).toBeNull()
  })
  it('separates providers from transport protocols', () => {
    expect(modelProvider('claude-example', 'openai')).toBe('anthropic')
    expect(modelProvider('glm-example', 'deepseek')).toBe('zhipu')
  })
  it('rejects invalid prices', () => {
    const m = model()
    m.pricing = { billing_mode: 'token', input_price: NaN } as NonNullable<
      PlazaModel['pricing']
    >
    expect(
      minimumTokenPrice(buildCatalog([group(1, [m])])[0], 'input_price')
    ).toBeNull()
  })
})
