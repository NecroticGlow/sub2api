import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { ref } from 'vue'
import WanwuHome from '../WanwuHome.vue'

const { getModelPlaza } = vi.hoisted(() => ({ getModelPlaza: vi.fn() }))
vi.mock('@/api/modelPlaza', () => ({ getModelPlaza }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ locale: ref('zh') })
}))

const props = {
  siteName: '万物Ai',
  siteLogo: '',
  docUrl: '',
  apiBaseUrl: 'https://wanwuplus.com',
  isAuthenticated: false,
  dashboardPath: '/dashboard',
  showModelPlazaEntry: true
}
const catalog = {
  groups: [
    {
      id: 1,
      name: 'Public',
      platform: 'openai',
      is_exclusive: false,
      rate_multiplier: 1,
      models: [
        {
          name: 'gpt-example',
          platform: 'openai',
          popularity_rank: 2,
          pricing: null
        },
        {
          name: 'deepseek-example',
          platform: 'deepseek',
          popularity_rank: 1,
          pricing: null
        }
      ]
    },
    {
      id: 2,
      name: 'Private',
      platform: 'openai',
      is_exclusive: true,
      rate_multiplier: 1,
      models: [
        {
          name: 'private-model',
          platform: 'openai',
          popularity_rank: 1,
          pricing: null
        }
      ]
    }
  ]
}
const wrappers: ReturnType<typeof mount>[] = []
function render(overrides = {}) {
  const wrapper = mount(WanwuHome, {
    props: { ...props, ...overrides },
    global: {
      stubs: { RouterLink: RouterLinkStub, LocaleSwitcher: true }
    }
  })
  wrappers.push(wrapper)
  return wrapper
}

beforeEach(() => {
  getModelPlaza.mockReset().mockResolvedValue(catalog)
  vi.spyOn(window, 'matchMedia').mockReturnValue({
    matches: false,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn()
  } as unknown as MediaQueryList)
})
afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  vi.restoreAllMocks()
})

describe('Wanwu homepage', () => {
  it('shows real public models when the backend does not supply popularity ranks', async () => {
    const unranked = structuredClone(catalog)
    for (const group of unranked.groups) {
      for (const model of group.models) Reflect.deleteProperty(model, 'popularity_rank')
    }
    getModelPlaza.mockResolvedValueOnce(unranked)
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('.popular-card h3').map((el) => el.text())).toEqual([
      'gpt-example', 'deepseek-example'
    ])
    expect(wrapper.text()).not.toContain('近 7 天')
    expect(wrapper.text()).not.toContain('private-model')
  })

  it('renders generated responsive art and links visitors to working destinations', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('h1')).toHaveLength(1)
    expect(wrapper.get('h1').text()).toContain('万物一体')
    expect(wrapper.findAll('.world-card img')).toHaveLength(3)
    expect(wrapper.get('.hero-art img').attributes('fetchpriority')).toBe(
      'high'
    )
    expect(
      wrapper
        .findAll('.world-card img')
        .every((img) => img.attributes('loading') === 'lazy')
    ).toBe(true)
    expect(
      wrapper
        .findAllComponents(RouterLinkStub)
        .some((link) => link.props('to') === '/login')
    ).toBe(true)
    expect(wrapper.findAll('.popular-card h3').map((el) => el.text())).toEqual([
      'deepseek-example',
      'gpt-example'
    ])
    expect(wrapper.text()).not.toContain('private-model')
  })

  it('does not fetch or show restricted model data when the plaza is hidden', async () => {
    const wrapper = render({ showModelPlazaEntry: false })
    await flushPromises()
    expect(getModelPlaza).not.toHaveBeenCalled()
    expect(wrapper.find('.popular-section').exists()).toBe(false)
    expect(
      wrapper
        .findAllComponents(RouterLinkStub)
        .some((link) => link.props('to') === '/model-plaza')
    ).toBe(false)
    expect(wrapper.get('pre').text()).toContain('YOUR_MODEL_ID')
  })

  it('uses an honest error state without fabricated model entries', async () => {
    getModelPlaza.mockRejectedValueOnce(new Error('offline'))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('.popular-card')).toHaveLength(0)
    expect(wrapper.get('.model-message').text()).toContain('暂时未能加载')
  })

  it('switches code examples, copies actual displayed code and reports clipboard failure', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText }
    })
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('pre').text()).toContain(
      'https://wanwuplus.com/v1/chat/completions'
    )
    expect(wrapper.get('pre').text()).toContain('$WANWU_API_KEY')
    await wrapper.findAll('.code-tabs button')[1].trigger('click')
    expect(wrapper.get('pre').text()).toContain('from openai import OpenAI')
    expect(wrapper.get('pre').text()).toContain('model="gpt-example"')
    await wrapper.get('.copy-code').trigger('click')
    expect(writeText).toHaveBeenCalledWith(wrapper.get('pre').text())
    expect(wrapper.get('.copy-code').text()).toContain('已复制')
    writeText.mockRejectedValueOnce(new Error('denied'))
    await wrapper.get('.copy-code').trigger('click')
    expect(wrapper.get('.copy-error').text()).toContain('手动复制')
  })

  it('opens and closes mobile navigation and routes signed-in administrators correctly', async () => {
    const wrapper = render({
      isAuthenticated: true,
      dashboardPath: '/admin/dashboard'
    })
    await wrapper.get('.menu-toggle').trigger('click')
    expect(wrapper.get('.menu-toggle').attributes('aria-expanded')).toBe('true')
    await wrapper.get('.mobile-nav a').trigger('click')
    expect(wrapper.find('.mobile-nav').exists()).toBe(false)
    expect(
      wrapper
        .findAllComponents(RouterLinkStub)
        .some((link) => link.props('to') === '/admin/dashboard')
    ).toBe(true)
  })

  it('respects reduced motion and disables the animation control', async () => {
    vi.spyOn(window, 'matchMedia').mockImplementation(
      (query) =>
        ({
          matches: query.includes('reduced-motion'),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn()
        }) as unknown as MediaQueryList
    )
    const wrapper = render()
    await flushPromises()
    expect(wrapper.classes()).toContain('motion-off')
    expect(wrapper.get('.motion-control').attributes('disabled')).toBeDefined()
    expect(wrapper.get('.motion-control').text()).toContain('已减少动态效果')
  })

  it('allows visitors to pause animation and aborts catalog loading when unmounted', async () => {
    const wrapper = render()
    const signal = getModelPlaza.mock.calls[0][0].signal as AbortSignal
    await wrapper.get('.motion-control').trigger('click')
    expect(wrapper.get('.motion-control').attributes('aria-pressed')).toBe(
      'true'
    )
    expect(wrapper.classes()).toContain('motion-off')
    wrapper.unmount()
    expect(signal.aborted).toBe(true)
  })
})
