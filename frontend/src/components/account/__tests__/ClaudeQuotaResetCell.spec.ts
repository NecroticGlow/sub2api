import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ClaudeQuotaResetCell from '../ClaudeQuotaResetCell.vue'
import type { Account } from '@/types'

const { refresh, reset } = vi.hoisted(() => ({ refresh: vi.fn(), reset: vi.fn() }))
vi.mock('@/api/admin/accounts', () => ({ refreshClaudeResetCredits: refresh, resetClaudeQuota: reset }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const status = () => ({ eligible: true, available_count: 1, next_grant_id: 'grant', fetched_at: new Date().toISOString(), grants: [{ id: 'grant', resets_left: 1, usable_now: true, paused: false, ends_at: new Date(Date.now() + 86400_000).toISOString() }] })
const account = (snapshot?: unknown) => ({ id: 42, platform: 'anthropic', type: 'oauth', extra: { claude_reset_credit_snapshot: snapshot } } as unknown as Account)
const options = { global: { stubs: { ConfirmDialog: { props: ['show'], template: '<button v-if="show" data-test="confirm" @click="$emit(\'confirm\')">confirm</button>' } } } }

beforeEach(() => vi.clearAllMocks())
describe('Claude reset credits', () => {
  it('reads server snapshot without querying or consuming on mount', () => {
    const wrapper = mount(ClaudeQuotaResetCell, { ...options, props: { account: account(status()) } })
    expect(wrapper.text()).toContain('count 1')
    expect(refresh).not.toHaveBeenCalled()
    expect(reset).not.toHaveBeenCalled()
    expect(wrapper.get('[data-test="reset"]').attributes('disabled')).toBeUndefined()
  })
  it('requires query and confirmation before consumption', async () => {
    refresh.mockResolvedValue(status())
    reset.mockResolvedValue({ result: 'reset', credits: null, refresh_warning: false })
    const wrapper = mount(ClaudeQuotaResetCell, { ...options, props: { account: account() } })
    expect(wrapper.get('[data-test="reset"]').attributes('disabled')).toBeDefined()
    await wrapper.get('[data-test="query"]').trigger('click'); await flushPromises()
    expect(reset).not.toHaveBeenCalled()
    await wrapper.get('[data-test="reset"]').trigger('click')
    expect(reset).not.toHaveBeenCalled()
    await wrapper.get('[data-test="confirm"]').trigger('click'); await flushPromises()
    expect(reset).toHaveBeenCalledWith(42, { confirmed: true, grant_id: 'grant', resets_left: 1 })
    expect(wrapper.text()).toContain('results.reset')
    expect(wrapper.emitted('reset-completed')).toHaveLength(1)
  })
  it('disables expired, paused, stale and ineligible grants', () => {
    const cases = [
      { ...status(), eligible: false },
      { ...status(), fetched_at: new Date(Date.now() - 25 * 3600_000).toISOString() },
      { ...status(), grants: [{ ...status().grants[0]!, paused: true }] },
      { ...status(), grants: [{ ...status().grants[0]!, ends_at: '2020-01-01T00:00:00Z' }] }
    ]
    for (const snapshot of cases) {
      const wrapper = mount(ClaudeQuotaResetCell, { ...options, props: { account: account(snapshot) } })
      expect(wrapper.get('[data-test="reset"]').attributes('disabled')).toBeDefined()
    }
  })
  it('ambiguous POST clears actionable state instead of automatic retry', async () => {
    reset.mockRejectedValue(new Error('synthetic timeout'))
    const wrapper = mount(ClaudeQuotaResetCell, { ...options, props: { account: account(status()) } })
    await wrapper.get('[data-test="reset"]').trigger('click')
    await wrapper.get('[data-test="confirm"]').trigger('click'); await flushPromises()
    expect(reset).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-test="reset"]').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('outcomeUnknown')
  })
  it('does not apply a late query to a different account', async () => {
    let resolve!: (value: unknown) => void
    refresh.mockReturnValue(new Promise(r => { resolve = r }))
    const wrapper = mount(ClaudeQuotaResetCell, { ...options, props: { account: account() } })
    await wrapper.get('[data-test="query"]').trigger('click')
    await wrapper.setProps({ account: { ...account(), id: 43 } })
    resolve(status()); await flushPromises()
    expect(wrapper.text()).not.toContain('count 1')
  })
})
