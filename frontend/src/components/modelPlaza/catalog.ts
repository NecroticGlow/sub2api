import type { ModelPlazaGroup, PlazaModel } from '@/api/modelPlaza'

export interface CatalogRoute {
  group: ModelPlazaGroup
  model: PlazaModel
}
export interface CatalogModel {
  key: string
  name: string
  platform: string
  routes: CatalogRoute[]
  popularityRank: number
}

/** Keep provider identity separate from the transport platform (e.g. Claude over OpenAI). */
export function modelProvider(name: string, platform: string): string {
  const value = name.toLowerCase()
  if (value.includes('claude')) return 'anthropic'
  if (value.includes('gemini')) return 'gemini'
  if (value.includes('deepseek')) return 'deepseek'
  if (value.includes('grok')) return 'grok'
  if (value.includes('glm')) return 'zhipu'
  if (/^(gpt-|o[134](?:-|$)|codex)/.test(value)) return 'openai'
  return platform
}

export function buildCatalog(groups: ModelPlazaGroup[]): CatalogModel[] {
  const catalog = new Map<string, CatalogModel>()
  for (const group of groups) {
    for (const model of group.models) {
      // A model's API protocol is part of its identity; never mix incompatible routes.
      const key = `${model.platform}:${model.name.toLowerCase()}`
      const item: CatalogModel = catalog.get(key) ?? {
        key,
        name: model.name,
        platform: model.platform,
        routes: [],
        popularityRank: model.popularity_rank ?? Infinity
      }
      item.popularityRank = Math.min(item.popularityRank, model.popularity_rank ?? Infinity)
      if (!item.routes.some((route) => route.group.id === group.id))
        item.routes.push({ group, model })
      catalog.set(key, item)
    }
  }
  return [...catalog.values()]
}

export type TokenPriceField =
  | 'input_price'
  | 'output_price'
  | 'cache_read_price'
  | 'cache_write_price'
  | 'cache_write_1h_price'

/** Standard token prices only. Unknown and non-token prices must never appear as free. */
export function tokenPrice(
  route: CatalogRoute,
  field: TokenPriceField
): number | null {
  const pricing = route.model.pricing
  if (!pricing || (pricing.billing_mode && pricing.billing_mode !== 'token'))
    return null
  const value = pricing[field]
  const rate = route.group.user_rate_multiplier ?? route.group.rate_multiplier
  if (
    value == null ||
    !Number.isFinite(value) ||
    !Number.isFinite(rate) ||
    value < 0 ||
    rate < 0
  )
    return null
  return value * rate * 1_000_000
}

export function minimumTokenPrice(
  model: CatalogModel,
  field: TokenPriceField
): number | null {
  const prices = model.routes
    .map((route) => tokenPrice(route, field))
    .filter((v): v is number => v !== null)
  return prices.length ? Math.min(...prices) : null
}

export function filterCatalog(
  models: CatalogModel[],
  provider: string,
  query: string
): CatalogModel[] {
  const words = query.trim().toLowerCase().split(/\s+/).filter(Boolean)
  return models.filter((model) => {
    if (
      provider !== 'all' &&
      modelProvider(model.name, model.platform) !== provider
    )
      return false
    const haystack =
      `${model.name} ${model.platform} ${modelProvider(model.name, model.platform)} ${model.routes.map((r) => r.group.name).join(' ')}`.toLowerCase()
    return words.every((word) => haystack.includes(word))
  })
}
