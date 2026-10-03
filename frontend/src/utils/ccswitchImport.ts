import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.6-sol'
export const GROK_CC_SWITCH_MODEL = 'grok-4.6'
export const ANTHROPIC_CC_SWITCH_MODEL = 'claude-opus-4-8'
export const DEEPSEEK_CC_SWITCH_MODEL = 'deepseek-chat'
export const CC_SWITCH_PROVIDER_API_BASE_URL = 'https://wanwuplus.com'

export type CcSwitchApp = 'claude' | 'codex' | 'gemini' | 'grokbuild' | 'opencode'

export interface CcSwitchImportDeeplinkInput {
  homepage: string
  platform?: GroupPlatform | null
  app: CcSwitchApp
  providerName: string
  apiKey: string
  usageScript: string
  model?: string
  haikuModel?: string
  sonnetModel?: string
  opusModel?: string
}

/**
 * Balance query CC Switch runs against the imported provider. CC Switch fills
 * `{{baseUrl}}` with the provider's base URL as stored — Codex and Grok imports
 * carry a trailing `/v1` (see `withV1Endpoint`), Claude ones do not, and users
 * may edit it either way afterwards — then evaluates the script, so the URL
 * strips an existing `/v1` instead of blindly appending one (`/v1/v1/usage`
 * is a 404 and CC Switch shows "query failed").
 */
export const CC_SWITCH_USAGE_SCRIPT = `({
    request: {
      url: "{{baseUrl}}".replace(/\\/+$/, "").replace(/\\/v1$/, "") + "/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

export function defaultCcSwitchAppForPlatform(platform: GroupPlatform | undefined | null): CcSwitchApp {
  switch (platform || 'anthropic') {
    case 'openai': return 'codex'
    case 'gemini': return 'gemini'
    case 'grok': return 'grokbuild'
    default: return 'claude'
  }
}

export function defaultCcSwitchModelForPlatform(platform: GroupPlatform | undefined | null): string {
  switch (platform || 'anthropic') {
    case 'openai': return OPENAI_CC_SWITCH_CODEX_MODEL
    case 'grok': return GROK_CC_SWITCH_MODEL
    case 'anthropic': return ANTHROPIC_CC_SWITCH_MODEL
    case 'deepseek': return DEEPSEEK_CC_SWITCH_MODEL
    default: return ''
  }
}

export function resolveCcSwitchEndpoint(platform: GroupPlatform | undefined | null, baseUrl: string): string {
  switch (platform || 'anthropic') {
    case 'antigravity': return `${baseUrl.replace(/\/+$/, '')}/antigravity`
    case 'openai': return baseUrl.replace(/\/+$/, '')
    case 'grok': return withV1Endpoint(baseUrl)
    default: return baseUrl
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const endpoint = input.app === 'opencode'
    ? withV1Endpoint(CC_SWITCH_PROVIDER_API_BASE_URL)
    : resolveCcSwitchEndpoint(input.platform, CC_SWITCH_PROVIDER_API_BASE_URL)
  const homepage = input.homepage.trim().replace(/\/+$/, '')
  const entries: [string, string][] = [
    ['resource', 'provider'], ['app', input.app], ['name', input.providerName],
    ['homepage', homepage], ['endpoint', endpoint], ['apiKey', input.apiKey],
    ['configFormat', 'json'], ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)], ['usageAutoInterval', '30']
  ]
  const model = input.model?.trim()
  if (model) entries.splice(2, 0, ['model', model])
  if (input.app === 'claude') {
    for (const [key, value] of [['haikuModel', input.haikuModel], ['sonnetModel', input.sonnetModel], ['opusModel', input.opusModel]] as const) {
      const trimmed = value?.trim()
      if (trimmed) entries.push([key, trimmed])
    }
  }
  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}

export function ccSwitchModelsUrls(baseUrl: string, currentOrigin: string, platform?: GroupPlatform | null): string[] {
  const urls = new Set<string>()
  for (const candidate of [currentOrigin, baseUrl, CC_SWITCH_PROVIDER_API_BASE_URL]) {
    try {
      const url = new URL(candidate.trim())
      if (url.protocol !== 'https:' && url.protocol !== 'http:') continue
      url.search = ''
      url.hash = ''
      let endpoint = url.toString().replace(/\/+$/, '')
      if (platform === 'antigravity') {
        endpoint = endpoint.replace(/\/v1$/, '')
        if (!endpoint.endsWith('/antigravity')) endpoint += '/antigravity'
      }
      urls.add(withV1Endpoint(endpoint) + '/models')
    } catch { /* Ignore malformed settings and continue with a usable endpoint. */ }
  }
  return [...urls]
}

/** Accept both the OpenAI list and the Codex model catalogue without inventing IDs. */
export function parseCcSwitchModelIds(payload: unknown): string[] {
  if (!payload || typeof payload !== 'object') return []
  const body = payload as { data?: unknown; models?: unknown }
  const items = Array.isArray(body.data) ? body.data : body.models
  if (!Array.isArray(items)) return []
  const models = new Set<string>()
  for (const item of items) {
    let id: unknown
    if (typeof item === 'string') {
      id = item
    } else if (item && typeof item === 'object') {
      const model = item as { id?: unknown; slug?: unknown; name?: unknown }
      id = model.id ?? model.slug ?? model.name
      if (id === model.name && typeof id === 'string') id = id.replace(/^models\//, '')
    }
    if (typeof id === 'string' && id.trim()) models.add(id.trim())
  }
  return [...models]
}
