import { describe, expect, it } from 'vitest'
import {
  ANTHROPIC_CC_SWITCH_MODEL,
  DEEPSEEK_CC_SWITCH_MODEL,
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  buildCcSwitchImportDeeplink,
  ccSwitchModelsUrls,
  defaultCcSwitchAppForPlatform,
  defaultCcSwitchModelForPlatform,
  resolveCcSwitchEndpoint
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

describe('ccswitchImport utils', () => {
  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.6-sol')
    expect(defaultCcSwitchModelForPlatform('openai')).toBe('gpt-5.6-sol')
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.6')
    expect(defaultCcSwitchModelForPlatform('grok')).toBe('grok-4.6')
  })

  it('defaults Anthropic imports to the current Opus model', () => {
    expect(ANTHROPIC_CC_SWITCH_MODEL).toBe('claude-opus-4-8')
    expect(defaultCcSwitchModelForPlatform('anthropic')).toBe('claude-opus-4-8')
    expect(defaultCcSwitchModelForPlatform(null)).toBe('claude-opus-4-8')
  })

  it('defaults DeepSeek imports to deepseek-chat', () => {
    expect(DEEPSEEK_CC_SWITCH_MODEL).toBe('deepseek-chat')
    expect(defaultCcSwitchModelForPlatform('deepseek')).toBe('deepseek-chat')
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, app: 'claude' },
    { platform: 'openai' as GroupPlatform, app: 'codex' },
    { platform: 'gemini' as GroupPlatform, app: 'gemini' },
    { platform: 'antigravity' as GroupPlatform, app: 'claude' },
    { platform: 'grok' as GroupPlatform, app: 'grokbuild' },
    { platform: 'deepseek' as GroupPlatform, app: 'claude' },
    { platform: null, app: 'claude' }
  ])('defaults $platform imports to the $app app', ({ platform, app }) => {
    expect(defaultCcSwitchAppForPlatform(platform)).toBe(app)
  })

  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it('adds the selected model parameter for OpenAI imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        app: 'codex',
        model: OPENAI_CC_SWITCH_CODEX_MODEL
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it('imports OpenCode with an OpenAI-compatible /v1 endpoint', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        app: 'opencode',
        model: 'gpt-5.6-sol'
      })
    )

    expect(params.get('app')).toBe('opencode')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe('gpt-5.6-sol')
  })

  it('loads models from same origin first and keeps a configured endpoint fallback', () => {
    expect(ccSwitchModelsUrls('https://api.example.com/', 'https://panel.example.com')).toEqual([
      'https://panel.example.com/v1/models',
      'https://api.example.com/v1/models'
    ])
    expect(ccSwitchModelsUrls('https://panel.example.com', 'https://panel.example.com')).toEqual([
      'https://panel.example.com/v1/models'
    ])
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        app: 'grokbuild',
        model: GROK_CC_SWITCH_MODEL
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
    expect(resolveCcSwitchEndpoint('grok', baseUrl)).toBe('https://api.example.com/v1')
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, app: 'claude' as const },
    { platform: 'gemini' as GroupPlatform, app: 'gemini' as const }
  ])('omits the model parameter when none is selected for $platform', ({ platform, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        app
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.has('model')).toBe(false)
  })

  it('keeps Antigravity imports on the antigravity endpoint', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        app: 'gemini'
      })
    )

    expect(params.get('app')).toBe('gemini')
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/antigravity`)
    expect(params.has('model')).toBe(false)
  })

  it('includes tiered Claude models only for the claude app', () => {
    const claudeParams = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'anthropic',
        app: 'claude',
        model: 'claude-sonnet-4-5',
        haikuModel: 'claude-haiku-4-5',
        sonnetModel: 'claude-sonnet-4-5',
        opusModel: 'claude-opus-4-6'
      })
    )

    expect(claudeParams.get('model')).toBe('claude-sonnet-4-5')
    expect(claudeParams.get('haikuModel')).toBe('claude-haiku-4-5')
    expect(claudeParams.get('sonnetModel')).toBe('claude-sonnet-4-5')
    expect(claudeParams.get('opusModel')).toBe('claude-opus-4-6')

    const codexParams = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        app: 'codex',
        model: 'gpt-5.5',
        haikuModel: 'ignored',
        sonnetModel: 'ignored',
        opusModel: 'ignored'
      })
    )

    expect(codexParams.has('haikuModel')).toBe(false)
    expect(codexParams.has('sonnetModel')).toBe(false)
    expect(codexParams.has('opusModel')).toBe(false)
  })

  it('trims whitespace-only model values', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'anthropic',
        app: 'claude',
        model: '   ',
        haikuModel: '  '
      })
    )

    expect(params.has('model')).toBe(false)
    expect(params.has('haikuModel')).toBe(false)
  })
})
