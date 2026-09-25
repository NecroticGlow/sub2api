import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import IntelligenceTestView from '../IntelligenceTestView.vue'

const { list, testIntelligence } = vi.hoisted(() => ({ list: vi.fn(), testIntelligence: vi.fn() }))
vi.mock('@/api/admin/accounts', () => ({ list, default: { list } }))
vi.mock('@/api/admin/intelligence', () => ({ testIntelligence }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

const account = (id: number, platform = 'openai', type = 'oauth') => ({ id, name: `Account ${id}`, platform, type, status: 'active', current_concurrency: id, concurrency: 5 })
const result = (id: number, status = 'manual_review') => ({ account_id: id, status, model: 'gpt-6-astra', response_text: '<script>unsafe</script> uncertain', checks: [{ item: 'iphone', expected: 'iPhone 17', matched: false }], tested_at: '2026-09-24T00:00:00Z', latency_ms: 1500 })
const mountView = () => shallowMount(IntelligenceTestView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' } } } })
let wrapper: ReturnType<typeof mountView>
beforeEach(() => {
  vi.clearAllMocks()
  list.mockResolvedValue({ items: [account(1), account(2), account(3, 'anthropic'), account(4, 'openai', 'apikey')], total: 2, pages: 1 })
  testIntelligence.mockResolvedValue(result(1))
})
afterEach(() => wrapper?.unmount())

describe('IntelligenceTestView', () => {
  it('requests only OpenAI OAuth accounts and defensively excludes other account types', async () => {
    wrapper = mountView()
    await flushPromises()
    expect(list).toHaveBeenCalledWith(1, 20, expect.objectContaining({ platform: 'openai', type: 'oauth', lite: 'true' }), { signal: expect.any(AbortSignal) })
    expect(wrapper.findAll('article')).toHaveLength(2)
    expect(wrapper.get('[data-account-id="1"]').text()).toContain('admin.intelligence.concurrency: 1 / 5')
    expect(wrapper.get('[data-account-id="1"]').text()).toContain('admin.intelligence.lastTest: admin.intelligence.idle')
    expect(testIntelligence).not.toHaveBeenCalled()
  })

  it('keeps tests independent, blocks double clicks and escapes raw answers', async () => {
    let finish!: (value: ReturnType<typeof result>) => void
    testIntelligence.mockReturnValue(new Promise(resolve => { finish = resolve }))
    wrapper = mountView()
    await flushPromises()
    const first = wrapper.get('[data-account-id="1"]')
    await first.get('button').trigger('click')
    expect(first.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-account-id="2"] button').attributes('disabled')).toBeUndefined()
    expect(testIntelligence).toHaveBeenCalledTimes(1)
    expect(testIntelligence).toHaveBeenCalledWith(1, expect.any(AbortSignal))
    finish(result(1))
    await flushPromises()
    expect(first.text()).toContain('admin.intelligence.manual_review')
    expect(first.get('pre').text()).toContain('<script>unsafe</script>')
    expect(first.find('script').exists()).toBe(false)
    const normal = first.findAll('button').find(button => button.text() === 'admin.intelligence.normal')!
    await normal.trigger('click')
    expect(first.text()).toContain('admin.intelligence.reviewed_normal')
    const degraded = first.findAll('button').find(button => button.text() === 'admin.intelligence.degraded')!
    await degraded.trigger('click')
    expect(first.text()).toContain('admin.intelligence.reviewed_degraded')
  })

  it('shows matched results and keeps failed requests separate from degradation judgments', async () => {
    testIntelligence.mockResolvedValueOnce(result(1, 'passed')).mockRejectedValueOnce(new Error('Upstream timeout'))
    wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-account-id="1"] button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-account-id="1"]').text()).toContain('admin.intelligence.passed')
    await wrapper.get('[data-account-id="2"] button').trigger('click')
    await flushPromises()
    const second = wrapper.get('[data-account-id="2"]')
    expect(second.text()).toContain('admin.intelligence.error')
    expect(second.text()).toContain('Upstream timeout')
    expect(second.text()).not.toContain('admin.intelligence.reviewed_degraded')
  })

  it('paginates with the OAuth filter and aborts outstanding work on exit', async () => {
    list.mockResolvedValue({ items: [account(1)], total: 22, pages: 2 })
    testIntelligence.mockReturnValue(new Promise(() => {}))
    wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-account-id="1"] button').trigger('click')
    const signal = testIntelligence.mock.lastCall![1] as AbortSignal
    const next = wrapper.findAll('button').find(button => button.text() === 'pagination.next')!
    await next.trigger('click')
    await flushPromises()
    expect(list.mock.lastCall?.[0]).toBe(2)
    expect(list.mock.lastCall?.[2]).toMatchObject({ platform: 'openai', type: 'oauth' })
    wrapper.unmount()
    expect(signal.aborted).toBe(true)
  })
})
