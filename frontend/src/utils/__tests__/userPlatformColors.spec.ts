import { describe, expect, it } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import * as raw from '../platformColors'
import * as user from '../userPlatformColors'
import UserGroupBadge from '@/components/common/UserGroupBadge.vue'
import UserPlatformIcon from '@/components/common/UserPlatformIcon.vue'

describe('user-only OpenCode branding', () => {
  it('uses DeepSeek label and all DeepSeek color variants without changing admin branding', () => {
    expect(user.platformLabel('opencode_go')).toBe('DeepSeek')
    for (const [name, fn] of Object.entries(user)) {
      if (name.startsWith('platform') && name !== 'platformLabel') {
        expect(fn('opencode_go'), name).toBe((raw[name as keyof typeof raw] as (p: string) => string)('deepseek'))
      }
    }
    expect(raw.platformLabel('opencode_go')).toMatch(/OpenCode/i)
    expect(raw.platformBadgeClass('opencode_go')).toContain('amber')
    expect(user.platformBadgeClass('openai')).toBe(raw.platformBadgeClass('openai'))
  })

  it('sanitizes display names only, leaving backend platform and model IDs intact', () => {
    const group = { id: 42, name: 'OpenCode GO Premium', platform: 'opencode_go', model: 'deepseek-v4.1-flash' }
    expect(user.userGroupName(group.name)).toBe('DeepSeek Premium')
    expect(user.userGroupName('opencode_go')).toBe('DeepSeek')
    expect(user.userGroupName('deepseek-f')).toBe('deepseek-f')
    expect(group.platform).toBe('opencode_go')
    expect(group.model).toBe('deepseek-v4.1-flash')
    expect(user.userPlatform(undefined)).toBeUndefined()
  })

  it('forwards group ID-independent attributes and DeepSeek presentation to the shared badge', () => {
    const wrapper = shallowMount(UserGroupBadge, {
      props: { name: 'OpenCode', platform: 'opencode_go' },
      attrs: { rateMultiplier: 0.25, showRate: true, subscriptionType: 'standard' },
      global: { stubs: { GroupBadge: { name: 'GroupBadge', props: ['name', 'platform', 'rateMultiplier', 'showRate', 'subscriptionType'], template: '<span />' } } }
    })
    expect(wrapper.findComponent({ name: 'GroupBadge' }).props()).toMatchObject({
      name: 'DeepSeek', platform: 'deepseek', rateMultiplier: 0.25, showRate: true, subscriptionType: 'standard'
    })
  })

  it('renders the DeepSeek icon without changing the shared account icon', () => {
    const wrapper = shallowMount(UserPlatformIcon, { props: { platform: 'opencode_go', size: 'lg' } })
    expect(wrapper.findComponent({ name: 'PlatformIcon' }).props()).toMatchObject({ platform: 'deepseek', size: 'lg' })
  })
})
