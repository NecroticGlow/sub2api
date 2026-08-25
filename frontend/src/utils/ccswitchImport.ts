import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.6-sol'
export const GROK_CC_SWITCH_MODEL = 'grok-4.6'
export const ANTHROPIC_CC_SWITCH_MODEL = 'claude-opus-4-8'
export const DEEPSEEK_CC_SWITCH_MODEL = 'deepseek-chat'

export type CcSwitchApp = 'claude' | 'codex' | 'gemini' | 'grokbuild' | 'opencode'

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
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

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

export function defaultCcSwitchAppForPlatform(
  platform: GroupPlatform | undefined | null
): CcSwitchApp {
  switch (platform || 'anthropic') {
    case 'openai': return 'codex'
    case 'gemini': return 'gemini'
    case 'grok': return 'grokbuild'
    default: return 'claude'
  }
}

export function defaultCcSwitchModelForPlatform(
  platform: GroupPlatform | undefined | null
): string {
  switch (platform || 'anthropic') {
    case 'openai': return OPENAI_CC_SWITCH_CODEX_MODEL
    case 'grok': return GROK_CC_SWITCH_MODEL
    case 'anthropic': return ANTHROPIC_CC_SWITCH_MODEL
    case 'deepseek': return DEEPSEEK_CC_SWITCH_MODEL
    default: return ''
  }
}

export function resolveCcSwitchEndpoint(
  platform: GroupPlatform | undefined | null,
  baseUrl: string
): string {
  switch (platform || 'anthropic') {
    case 'antigravity': return `${baseUrl}/antigravity`
    case 'grok': return withV1Endpoint(baseUrl)
    default: return baseUrl
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const endpoint = input.app === 'opencode'
    ? withV1Endpoint(input.baseUrl)
    : resolveCcSwitchEndpoint(input.platform, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'], ['app', input.app], ['name', input.providerName],
    ['homepage', input.baseUrl], ['endpoint', endpoint], ['apiKey', input.apiKey],
    ['configFormat', 'json'], ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)], ['usageAutoInterval', '30']
  ]

  const model = input.model?.trim()
  if (model) entries.splice(2, 0, ['model', model])

  if (input.app === 'claude') {
    const tiered: [string, string | undefined][] = [
      ['haikuModel', input.haikuModel], ['sonnetModel', input.sonnetModel],
      ['opusModel', input.opusModel]
    ]
    for (const [key, value] of tiered) {
      const trimmed = value?.trim()
      if (trimmed) entries.push([key, trimmed])
    }
  }
  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}

export function ccSwitchModelsUrls(baseUrl: string, currentOrigin: string): string[] {
  const urls = [withV1Endpoint(currentOrigin) + '/models']
  const configured = withV1Endpoint(baseUrl) + '/models'
  if (!urls.includes(configured)) urls.push(configured)
  return urls
}
