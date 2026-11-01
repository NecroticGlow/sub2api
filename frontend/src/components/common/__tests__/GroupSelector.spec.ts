import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import GroupSelector from '../GroupSelector.vue'

const authState = { isSimpleMode: false }

vi.mock('@/stores', () => ({ useAuthStore: () => authState }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const groups = [
  { id: 1, name: 'Basic', platform: 'anthropic', status: 'active' },
  { id: 2, name: 'Composite', platform: 'composite', status: 'active' }
] as any

const mountSelector = (modelValue: number[] = []) => mount(GroupSelector, {
  props: { modelValue, groups },
  global: { stubs: { GroupBadge: { props: ['name'], template: '<span>{{ name }}</span>' }, Icon: true } }
})

describe('GroupSelector simple-mode binding policy', () => {
  beforeEach(() => { authState.isSimpleMode = false })

  it('hides composite groups in simple mode and preserves basic groups', () => {
    authState.isSimpleMode = true
    const wrapper = mountSelector()
    expect(wrapper.text()).toContain('Basic')
    expect(wrapper.text()).not.toContain('Composite')
  })

  it('keeps composite groups available in advanced mode', () => {
    const wrapper = mountSelector()
    expect(wrapper.text()).toContain('Composite')
  })

  it('allows OpenCode accounts in DeepSeek groups without opening other provider bindings', () => {
    const options = [
      { id: 10, name: 'DeepSeek group', platform: 'deepseek', status: 'active' },
      { id: 11, name: 'OpenCode group', platform: 'opencode_go', status: 'active' },
      { id: 12, name: 'GPT group', platform: 'openai', status: 'active' },
    ] as any
    const wrapper = mount(GroupSelector, { props: { modelValue: [], groups: options, platform: 'opencode_go' },
      global: { stubs: { GroupBadge: { props: ['name'], template: '<span>{{ name }}</span>' }, Icon: true } } })
    expect(wrapper.text()).toContain('DeepSeek group')
    expect(wrapper.text()).toContain('OpenCode group')
    expect(wrapper.text()).not.toContain('GPT group')
    wrapper.unmount()
    const native = mount(GroupSelector, { props: { modelValue: [], groups: options, platform: 'deepseek' },
      global: { stubs: { GroupBadge: { props: ['name'], template: '<span>{{ name }}</span>' }, Icon: true } } })
    expect(native.text()).not.toContain('OpenCode group')
    native.unmount()
  })

  it('cleans hidden historical composite IDs while preserving visible selections', () => {
    authState.isSimpleMode = true
    const wrapper = mountSelector([1, 2])
    expect(wrapper.emitted('update:modelValue')).toEqual([[[1]]])
  })
})
