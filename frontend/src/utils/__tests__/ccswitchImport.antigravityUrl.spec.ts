import { describe, expect, it } from 'vitest'
import { buildCcSwitchImportDeeplink, resolveCcSwitchEndpoint, CC_SWITCH_PROVIDER_API_BASE_URL } from '../ccswitchImport'

describe('Antigravity CC Switch endpoint', () => {
  it.each([
    ['https://api.example.com', 'https://api.example.com/antigravity'],
    ['https://api.example.com/', 'https://api.example.com/antigravity'],
    ['https://api.example.com///', 'https://api.example.com/antigravity'],
    ['https://api.example.com/sub2api/', 'https://api.example.com/sub2api/antigravity']
  ])('joins the platform path onto %s for both clients', (baseUrl, endpoint) => {
    expect(resolveCcSwitchEndpoint('antigravity', baseUrl)).toBe(endpoint)
    for (const clientType of ['claude', 'gemini'] as const) {
      const url = new URL(buildCcSwitchImportDeeplink({
        homepage: baseUrl, app: clientType, platform: 'antigravity', providerName: 'Sub2API',
        apiKey: 'sk-test', usageScript: 'return true'
      }))
      expect(url.searchParams.get('endpoint')).toBe(`${CC_SWITCH_PROVIDER_API_BASE_URL}/antigravity`)
      expect(url.searchParams.get('app')).toBe(clientType)
      expect(url.searchParams.get('homepage')).toBe(baseUrl.replace(/\/+$/, ''))
      expect(url.searchParams.has('model')).toBe(false)
    }
  })
})
