import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { InternalAxiosRequestConfig } from 'axios'
import { apiClient } from '@/api/client'
import OpenAIAuthConnections from '../OpenAIAuthConnections.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/i18n', () => ({ getLocale: () => 'en' }))
const originalAdapter = apiClient.defaults.adapter
const adapter = vi.fn()
const record = (id: string, kind = 'siwc', capabilities: string[] = []) => ({ id, account_id: 42, kind, label: `Synthetic ${id}`, legacy: false, status: 'active', client_id: `issued-${id}`, scopes: [], capabilities, preferred: false, revision: 2 })
let records: ReturnType<typeof record>[]
beforeEach(() => {
  vi.resetAllMocks(); localStorage.clear(); records = [record('identity'), record('responses', 'siwc', ['responses'])]
  adapter.mockImplementation(async (config: InternalAxiosRequestConfig) => ({ config, status: 200, statusText: 'OK', headers: {}, data: { code: 0, message: '', data: config.method === 'get' ? records : config.url?.endsWith('/authorize') ? { session_id: 'synthetic-session', auth_url: 'https://auth.openai.com/authorize' } : {} } }))
  apiClient.defaults.adapter = adapter
})
afterEach(() => { apiClient.defaults.adapter = originalAdapter })

describe('independent OpenAI connection management through actual API client', () => {
  it('does not nest forms or submit the surrounding account editor', async () => {
    const submit = vi.fn()
    const w = mount({ components: { OpenAIAuthConnections }, template: '<form @submit.prevent="submit"><OpenAIAuthConnections :account-id="42" :active="true" /></form>', setup: () => ({ submit }) })
    await flushPromises()
    expect(w.findAll('form')).toHaveLength(1)
    await w.get('[data-testid="connection-label"]').setValue('New registration')
    await w.get('[data-testid="add-connection"]').trigger('click'); await flushPromises()
    expect(submit).not.toHaveBeenCalled(); w.unmount()
  })

  it('shows both registrations and refuses identity-only selection', async () => {
    const w = mount(OpenAIAuthConnections, { props: { accountId: 42, active: true } }); await flushPromises()
    expect(w.text()).toContain('Synthetic identity'); expect(w.text()).toContain('Synthetic responses')
    const selects = w.findAll('[data-testid="select-connection"]')
    expect(selects[0].attributes('disabled')).toBeDefined(); expect(selects[1].attributes('disabled')).toBeUndefined()
    await selects[1].trigger('click'); await flushPromises()
    const request = adapter.mock.calls.map(([c]) => c).find(c => c.method === 'patch')
    expect(request.url).toBe('/admin/accounts/42/auth-connections/responses')
    expect(JSON.parse(request.data)).toEqual({ revision: 2, action: 'select' }); w.unmount()
  })
  it('creates a new registration without replacing existing ones', async () => {
    const w = mount(OpenAIAuthConnections, { props: { accountId: 42, active: true } }); await flushPromises()
    await w.get('[data-testid="connection-label"]').setValue('Second Codex')
    await w.get('[data-testid="add-connection"]').trigger('click'); await flushPromises()
    const request = adapter.mock.calls.map(([c]) => c).find(c => c.method === 'post')
    expect(JSON.parse(request.data)).toEqual({ kind: 'codex', label: 'Second Codex' })
    expect(w.text()).toContain('Synthetic identity'); w.unmount()
  })
  it('clears authorization material when the account or dialog changes', async () => {
    const w = mount(OpenAIAuthConnections, { props: { accountId: 42, active: true } }); await flushPromises()
    const button = w.findAll('button').find(b => b.text() === 'openaiAuth.authorize')!
    await button.trigger('click'); await flushPromises(); expect(w.find('[data-testid="connection-authorization"]').exists()).toBe(true)
    await w.setProps({ accountId: 43 }); await flushPromises()
    expect(w.find('[data-testid="connection-authorization"]').exists()).toBe(false)
    await w.setProps({ active: false }); expect(w.text()).not.toContain('Synthetic identity'); w.unmount()
  })
  it('makes returning to legacy storage explicit', async () => {
    const w = mount(OpenAIAuthConnections, { props: { accountId: 42, active: true } }); await flushPromises()
    await w.get('[data-testid="use-legacy"]').trigger('click'); await flushPromises()
    expect(adapter.mock.calls.some(([c]) => c.url === '/admin/accounts/42/auth-connections/use-legacy' && c.method === 'post')).toBe(true)
    w.unmount()
  })
})
