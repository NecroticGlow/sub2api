import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.6-sol'
export const GROK_CC_SWITCH_MODEL = 'grok-4.6'
export const ANTHROPIC_CC_SWITCH_MODEL = 'claude-opus-4-8'
export const DEEPSEEK_CC_SWITCH_MODEL = 'deepseek-chat'

/**
 * CC Switch target applications supported by the ccswitch:// v1 deeplink.
 * `grokbuild` is only meaningful for grok-platform groups.
 */
export type CcSwitchApp = 'claude' | 'codex' | 'gemini' | 'grokbuild' | 'opencode'

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  app: CcSwitchApp
  providerName: string
  apiKey: string
  usageScript: string
  /** Main model (ccswitch `model` param). Empty = let the client use its default. */
  model?: string
  /** Claude-only tiered models (ccswitch haikuModel/sonnetModel/opusModel). */
  haikuModel?: string
  sonnetModel?: string
  opusModel?: string
}

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

/** Default CC Switch app for a group platform. */
export function defaultCcSwitchAppForPlatform(
  platform: GroupPlatform | undefined | null
): CcSwitchApp {
  switch (platform || 'anthropic') {
    case 'openai':
      return 'codex'
    case 'gemini':
      return 'gemini'
    case 'grok':
      return 'grokbuild'
    default:
      // anthropic / antigravity / deepseek / composite default to Claude Code
      return 'claude'
  }
}

/** Suggested main model for a group platform (may be empty). */
export function defaultCcSwitchModelForPlatform(
  platform: GroupPlatform | undefined | null
): string {
  switch (platform || 'anthropic') {
    case 'openai':
      return OPENAI_CC_SWITCH_CODEX_MODEL
    case 'grok':
      return GROK_CC_SWITCH_MODEL
    case 'anthropic':
      return ANTHROPIC_CC_SWITCH_MODEL
    case 'deepseek':
      return DEEPSEEK_CC_SWITCH_MODEL
    default:
      // antigravity / composite: model sets vary, let the fetched list decide
      return ''
  }
}

/** Gateway endpoint CC Switch should call for a given group platform. */
export function resolveCcSwitchEndpoint(
  platform: GroupPlatform | undefined | null,
  baseUrl: string
): string {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return `${baseUrl}/antigravity`
    case 'grok':
      return withV1Endpoint(baseUrl)
    default:
      return baseUrl
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  // OpenCode's @ai-sdk/openai-compatible provider expects the OpenAI-compatible
  // API root, including /v1. Other CC Switch clients retain their historical
  // endpoint shapes.
  const endpoint = input.app === 'opencode'
    ? withV1Endpoint(input.baseUrl)
    : resolveCcSwitchEndpoint(input.platform, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', input.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  const model = input.model?.trim()
  if (model) {
    entries.splice(2, 0, ['model', model])
  }

  // Tiered models are only recognized by CC Switch for the Claude app.
  if (input.app === 'claude') {
    const tiered: [string, string | undefined][] = [
      ['haikuModel', input.haikuModel],
      ['sonnetModel', input.sonnetModel],
      ['opusModel', input.opusModel]
    ]
    for (const [key, value] of tiered) {
      const trimmed = value?.trim()
      if (trimmed) {
        entries.push([key, trimmed])
      }
    }
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}

/**
 * Candidate URLs for loading the gateway model catalog in the browser.
 * Prefer the current origin so split public/API domains cannot make the picker
 * fail on CORS; keep the configured public endpoint as a deployment fallback.
 */
export function ccSwitchModelsUrls(baseUrl: string, currentOrigin: string): string[] {
  const urls = [withV1Endpoint(currentOrigin) + '/models']
  const configured = withV1Endpoint(baseUrl) + '/models'
  if (!urls.includes(configured)) urls.push(configured)
  return urls
}
