import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import ModelPlazaContent from '../ModelPlazaContent.vue'
import type { ModelPlazaResponse } from '@/api/modelPlaza'

vi.mock('vue-i18n', async () => ({
  ...(await vi.importActual<typeof import('vue-i18n')>('vue-i18n')),
  useI18n: () => ({ locale: ref('zh'), t: (key: string) => key })
}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ isAuthenticated: false, isAdmin: false })
}))

const response = {
  description: '',
  groups: [
    {
      id: 1,
      name: 'Demo',
      description: '',
      platform: 'openai',
      rate_multiplier: 1,
      models: [
        { name: 'gpt-example', platform: 'openai', pricing: null },
        { name: 'claude-example', platform: 'anthropic', pricing: null }
      ]
    }
  ]
} as ModelPlazaResponse
const render = (props = {}) =>
  mount(ModelPlazaContent, {
    props: { response, loading: false, ...props },
    global: {
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        PlazaGroupSection: true
      }
    }
  })

beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn()
  window.matchMedia = vi.fn().mockReturnValue({ matches: true })
})

describe('ModelPlazaContent', () => {
  it('retains upstream platform and effective user-rate filters in the custom catalog', async () => {
    const first = response.groups[0]
    const wrapper = render({ response: { ...response, groups: [
      { ...first, models: [first.models[0]], rate_multiplier: 1, user_rate_multiplier: 0.25 },
      { ...first, id: 2, platform: 'anthropic', models: [first.models[1]], rate_multiplier: 0.5 }
    ] } })
    await wrapper.get('[data-testid="rate-filter"]').setValue('0.25')
    expect(wrapper.findAll('.model-card')).toHaveLength(1)
    expect(wrapper.get('.model-card h3').text()).toBe('gpt-example')
    await wrapper.get('.results-meta button').trigger('click')
    await wrapper.get('[data-testid="platform-filter"]').setValue('anthropic')
    expect(wrapper.get('.model-card h3').text()).toBe('claude-example')
    await wrapper.get('button[aria-label="分组价格表"]').trigger('click')
    expect(wrapper.findComponent({ name: 'PlazaGroupSection' }).props('group').id).toBe(2)
  })
  it('lists every provided cache price in CNY and follows the selected group', async () => {
    const model = { ...response.groups[0].models[0], pricing: {
      billing_mode: 'token', input_price: 1e-6, output_price: 2e-6,
      cache_read_price: 2e-8, cache_write_price: 0, cache_write_1h_price: 2e-6, intervals: []
    } }
    const wrapper = render({ embedded: true, response: { ...response, groups: [
      { ...response.groups[0], rate_multiplier: 0.4, models: [model] },
      { ...response.groups[0], id: 2, name: 'Second group', rate_multiplier: 1,
        models: [{ ...model, pricing: { ...model.pricing, cache_write_price: null, cache_write_1h_price: null } }] }
    ] } })
    expect(wrapper.get('.cache-read-price').text()).toContain('¥0.008')
    expect(wrapper.get('.cache-write-price strong').text()).toBe('¥0')
    expect(wrapper.get('.cache-write-hour-price').text()).toContain('¥0.8')
    expect(wrapper.get('.cache-read-price').text()).toContain('缓存读取（命中）')
    await wrapper.get('select[aria-label="接入分组"]').setValue('2')
    expect(wrapper.get('.cache-read-price strong').text()).toBe('¥0.02')
    expect(wrapper.get('.cache-write-price strong').text()).toBe('—')
    expect(wrapper.find('.cache-write-hour-price').exists()).toBe(false)
  })

  it('shows the current relay starting price in CNY with the group rate applied once', () => {
    const wrapper = render({ embedded: true, response: { ...response, groups: [{
      ...response.groups[0], rate_multiplier: 0.4, models: [{
        ...response.groups[0].models[0], pricing: {
          billing_mode: 'token', input_price: 1e-6, output_price: 2e-6, intervals: []
        }
      }]
    }] } })
    expect(wrapper.get('.prices').text()).toContain('¥0.4')
    expect(wrapper.get('.prices').text()).toContain('¥0.8')
    expect(wrapper.get('.prices').text()).not.toContain('$')
    expect(wrapper.get('.unit-hint').text()).toContain('人民币')
  })

  it('defaults to the seven-day popularity ranking and preserves ranks when filtering', async () => {
    const wrapper = render({ embedded: true, response: { ...response, groups: [{
      ...response.groups[0], models: [
        { ...response.groups[0].models[0], name: 'gpt-z', popularity_rank: 1 },
        { ...response.groups[0].models[0], name: 'gpt-a', popularity_rank: 2 }
      ]
    }] } })
    expect(wrapper.get('.console-intro').text()).toContain('近 7 天')
    expect(wrapper.findAll('.model-card h3').map(m => m.text())).toEqual(['gpt-z', 'gpt-a'])
    expect(wrapper.findAll('.rank-badge').map(m => m.text())).toEqual(['TOP 1', 'TOP 2'])
    await wrapper.get('input[type="search"]').setValue('gpt-a')
    expect(wrapper.get('.rank-badge').text()).toBe('TOP 2')
    await wrapper.get('input[type="search"]').setValue('')
    await wrapper.get('select[aria-label="排序方式"]').setValue('name')
    expect(wrapper.findAll('.model-card h3').map(m => m.text())).toEqual(['gpt-a', 'gpt-z'])
  })
  it('shows the compact console directory when embedded and can refresh it', async () => {
    const wrapper = render({ embedded: true })
    expect(wrapper.find('.hero').exists()).toBe(false)
    expect(wrapper.find('.getting-started').exists()).toBe(false)
    expect(wrapper.get('#catalog-heading').text()).toBe('热门模型 TOP 10')
    expect(wrapper.get('.console-stats strong').text()).toBe('2')
    expect(wrapper.findAll('.model-card')).toHaveLength(2)
    await wrapper.get('.refresh-action').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
    await wrapper.setProps({ loading: true })
    expect(wrapper.get('.refresh-action').attributes('disabled')).toBeDefined()
    expect(wrapper.get('.console-stats strong').text()).toBe('—')
  })
  it('shows real model counts and unknown prices without fabricated catalog entries', () => {
    const wrapper = render()
    expect(wrapper.findAll('.model-card')).toHaveLength(2)
    expect(wrapper.find('.prices').text()).toContain('—')
    expect(wrapper.find('.hero-stats strong').text()).toBe('2')
  })
  it('filters models and clears an empty search', async () => {
    const wrapper = render()
    await wrapper.get('input[type="search"]').setValue('claude')
    expect(wrapper.findAll('.model-card')).toHaveLength(1)
    await wrapper.get('input[type="search"]').setValue('not-a-model')
    expect(wrapper.get('.empty-state h3').text()).toContain('没有找到')
    await wrapper.get('.empty-state button').trigger('click')
    expect(wrapper.findAll('.model-card')).toHaveLength(2)
  })
  it('opens selected model details and preserves the available group', async () => {
    const wrapper = render()
    await wrapper.get('.card-details').trigger('click')
    expect(wrapper.get('#model-details h3').text()).toBe('claude-example')
    expect(
      wrapper.findComponent({ name: 'PlazaGroupSection' }).props('group').models
    ).toHaveLength(1)
  })
  it('provides a retry action when the API fails', async () => {
    const wrapper = render({ response: null, error: true })
    await wrapper.get('[role="alert"] button').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
  })
  it('filters group routes in cards, details and the price table', async () => {
    const first = response.groups[0]
    const wrapper = render({ response: { ...response, groups: [first, {
      ...first, id: 2, name: 'GPT public', models: [first.models[0]]
    }] } })
    await wrapper.get('select[aria-label="接入分组"]').setValue('2')
    expect(wrapper.findAll('.model-card')).toHaveLength(1)
    expect(wrapper.get('.model-card').text()).toContain('gpt-example')
    await wrapper.get('.card-details').trigger('click')
    expect(wrapper.findAllComponents({ name: 'PlazaGroupSection' })).toHaveLength(1)
    expect(wrapper.findComponent({ name: 'PlazaGroupSection' }).props('group').id).toBe(2)
    await wrapper.get('button[aria-label="分组价格表"]').trigger('click')
    expect(wrapper.findAllComponents({ name: 'PlazaGroupSection' })).toHaveLength(1)
    expect(wrapper.findComponent({ name: 'PlazaGroupSection' }).props('group').id).toBe(2)
    await wrapper.get('.results-meta button').trigger('click')
    expect(wrapper.findAllComponents({ name: 'PlazaGroupSection' })).toHaveLength(2)
  })
  it('renders an honest empty state instead of sample models', () => {
    const wrapper = render({ response: { description: '', groups: [] } })
    expect(wrapper.findAll('.model-card')).toHaveLength(0)
    expect(wrapper.get('.empty-state').text()).toContain('暂无近期调用记录')
  })
  it('sanitizes administrator-provided markdown', () => {
    const wrapper = render({
      response: {
        ...response,
        description: '<img src="x" onerror="alert(1)"><script>alert(1)</script>'
      }
    })
    expect(wrapper.get('.plaza-description').html()).not.toContain('onerror')
    expect(wrapper.get('.plaza-description').html()).not.toContain('<script')
  })
})
