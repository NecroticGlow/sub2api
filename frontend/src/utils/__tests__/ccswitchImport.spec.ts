import { describe, expect, it } from 'vitest'
import {
  CC_SWITCH_PROVIDER_API_BASE_URL,
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  buildCcSwitchImportDeeplink
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

describe('ccswitchImport utils', () => {
  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.6-sol')
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.6')
  })

  it('pins the API endpoint to wanwuplus.com and follows the current site for homepage', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        homepage: 'https://current-site.example.com/',
        platform: 'openai',
        app: 'codex',
        providerName: 'Sub2API',
        apiKey: 'sk-test',
        usageScript: 'return true'
      })
    )

    expect(CC_SWITCH_PROVIDER_API_BASE_URL).toBe('https://wanwuplus.com')
    expect(params.get('homepage')).toBe('https://current-site.example.com')
    expect(params.get('endpoint')).toBe('https://wanwuplus.com')
  })

  const baseInput = {
    homepage: 'https://current-site.example.com',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it('adds the Codex model parameter for OpenAI imports', () => {
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
    expect(params.get('homepage')).toBe(baseInput.homepage)
    expect(params.get('endpoint')).toBe(CC_SWITCH_PROVIDER_API_BASE_URL)
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it('imports Grok Build with one /v1 suffix on the fixed API endpoint', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'grok',
        app: 'grokbuild',
        model: GROK_CC_SWITCH_MODEL
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://wanwuplus.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, app: 'claude' as const },
    { platform: 'gemini' as GroupPlatform, app: 'gemini' as const }
  ])('does not add a model parameter for $platform imports', ({ platform, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        app
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(CC_SWITCH_PROVIDER_API_BASE_URL)
    expect(params.has('model')).toBe(false)
  })

  it('keeps Antigravity imports on the selected client endpoint without a model parameter', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        app: 'gemini'
      })
    )

    expect(params.get('app')).toBe('gemini')
    expect(params.get('endpoint')).toBe(`${CC_SWITCH_PROVIDER_API_BASE_URL}/antigravity`)
    expect(params.has('model')).toBe(false)
  })

  it('imports OpenCode with /v1 on the fixed API endpoint', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        app: 'opencode',
        model: OPENAI_CC_SWITCH_CODEX_MODEL
      })
    )

    expect(params.get('homepage')).toBe(baseInput.homepage)
    expect(params.get('endpoint')).toBe('https://wanwuplus.com/v1')
  })
})
