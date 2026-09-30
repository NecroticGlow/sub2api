import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CNProviderQuotaCell from '../CNProviderQuotaCell.vue'
import UsageProgressBar from '../UsageProgressBar.vue'
import type { Account } from '@/types'

const { queryQuota } = vi.hoisted(() => ({
  queryQuota: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    cnProviders: { queryQuota }
  }
}))

// 保留 vue-i18n 真实导出：UsageProgressBar 依赖 @/utils/format → @/i18n，
// 其模块级 createI18n 需要真实 createI18n 存在。
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const account = {
  id: 7,
  platform: 'zhipu',
  type: 'apikey',
  credentials: { account_mode: 'coding' },
  extra: {
    zhipu_5h_used_percent: 0,
    zhipu_weekly_used_percent: 27,
    zhipu_5h_reset_at: '2026-08-18T12:30:00+08:00',
    zhipu_weekly_reset_at: '2026-08-22T00:00:00+08:00',
    zhipu_usage_updated_at: new Date().toISOString()
  }
} as Account

describe('CNProviderQuotaCell', () => {
  it('does not render or auto-query a ClinePass account with usage disabled', async () => {
    const disabled = {
      ...account, id: 9999, platform: 'deepseek',
      credentials: { account_mode: 'payg', base_url: 'https://api.cline.bot', clinepass_usage_enabled: false },
      extra: { deepseek_5h_used_percent: 40, deepseek_usage_updated_at: '2020-01-01T00:00:00Z' }
    } as Account
    const wrapper = mount(CNProviderQuotaCell, { props: { account: disabled } })
    await flushPromises()
    expect(wrapper.find('[data-test="cn-provider-quota"]').exists()).toBe(false)
    expect(queryQuota).not.toHaveBeenCalled()
  })
  beforeEach(() => {
    queryQuota.mockReset()
  })

  it('renders persistent ClinePass 5h/weekly/monthly windows on a DeepSeek payg account and queries manually', async () => {
    const clineAccount = {
      ...account, id: 1135, platform: 'deepseek',
      credentials: { account_mode: 'payg', base_url: 'https://api.cline.bot/api/v1' },
      extra: {
        deepseek_5h_used_percent: 2,
        deepseek_weekly_used_percent: 51,
        deepseek_monthly_used_percent: 50,
        deepseek_usage_updated_at: new Date().toISOString()
      }
    } as Account
    queryQuota.mockResolvedValue({ success: true, tiers: [{ window: '5h', used_percent: 4 }] })
    const wrapper = mount(CNProviderQuotaCell, { props: { account: clineAccount } })
    await flushPromises()
    expect(wrapper.get('[data-test="clinepass-label"]').text()).toBe('ClinePass')
    expect(wrapper.findAllComponents(UsageProgressBar).map(bar => bar.props('utilization'))).toEqual([2, 51, 50])
    expect(queryQuota).not.toHaveBeenCalled()
    await wrapper.get('[data-test="cn-provider-quota-probe"]').trigger('click')
    await flushPromises()
    expect(queryQuota).toHaveBeenCalledWith(1135)
    expect(wrapper.findAllComponents(UsageProgressBar)[0].props('utilization')).toBe(4)
  })

  it('does not show ClinePass windows for an unrelated DeepSeek account', async () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account: {
      ...account, platform: 'deepseek', credentials: { account_mode: 'payg', base_url: 'https://api.deepseek.com' }
    } as Account } })
    await flushPromises()
    expect(wrapper.find('[data-test="cn-provider-quota"]').exists()).toBe(false)
    expect(queryQuota).not.toHaveBeenCalled()
  })

  it('renders tier rows through the shared UsageProgressBar inside the account table cell', async () => {
    queryQuota.mockResolvedValue({
      success: true,
      tiers: [
        { window: '5h', used_percent: 0, reset_at: '2026-08-18T12:30:00+08:00' },
        { window: 'weekly', used_percent: 27, reset_at: '2026-08-22T00:00:00+08:00' }
      ]
    })
    const wrapper = mount(CNProviderQuotaCell, { props: { account } })

    const root = wrapper.get('[data-test="cn-provider-quota"]')
    expect(root.classes()).toContain('min-w-[220px]')

    // 新鲜快照：挂载即渲染条形图，不触发探测
    await flushPromises()
    expect(queryQuota).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('27%')

    // probe 按钮文案是动词 key（i18n mock 返回 key 本身），点击触发查询
    const probeButton = root.get('[data-test="cn-provider-quota-probe"]')
    expect(probeButton.text()).toBe('admin.accounts.cnProviders.probe')
    await probeButton.trigger('click')
    await flushPromises()
    expect(queryQuota).toHaveBeenCalledWith(account.id)

    // tier 行由 UsageProgressBar 渲染：数量、label/color/utilization/reset 逐行对齐
    expect(root.findAll('[data-test="cn-provider-quota-tier"]')).toHaveLength(2)
    const bars = root.findAllComponents(UsageProgressBar)
    expect(bars).toHaveLength(2)
    expect(bars[0].props('label')).toBe('admin.accounts.cnProviders.window5h')
    expect(bars[0].props('utilization')).toBe(0)
    expect(bars[0].props('color')).toBe('indigo')
    expect(bars[0].props('resetsAt')).toBe('2026-08-18T12:30:00+08:00')
    expect(bars[1].props('label')).toBe('admin.accounts.cnProviders.windowWeekly')
    expect(bars[1].props('utilization')).toBe(27)
    expect(bars[1].props('color')).toBe('emerald')
    expect(bars[1].props('resetsAt')).toBe('2026-08-22T00:00:00+08:00')
  })

  it('labels the refresh control with an explicit action verb, not a data caption', async () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account } })
    await flushPromises()

    // The snapshot is fresh (usage_updated_at = now): bars render without probing.
    expect(queryQuota).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('27%')

    // The control reads as an action ("query"), unlike the old noun label
    // ("5-hour window/weekly window") which looked like a passive caption.
    // The i18n mock returns the key itself.
    const probeButton = wrapper.get('[data-test="cn-provider-quota-probe"]')
    expect(probeButton.text()).toBe('admin.accounts.cnProviders.probe')

    await probeButton.trigger('click')
    await flushPromises()
    expect(queryQuota).toHaveBeenCalledWith(account.id)
  })
})
