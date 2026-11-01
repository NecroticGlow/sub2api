import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import CcSwitchImportModal from '../CcSwitchImportModal.vue'
import type { ApiKey } from '@/types'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string, params?: { detail?: string }) => `${key}${params?.detail ? `: ${params.detail}` : ''}` }) }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))
vi.mock('@/api/keys', () => ({ keysAPI: { getModels: async (id: number, options: unknown) => {
  const result = await fetchMock(id, options)
  if (!result.ok) throw { status: result.status }
  return result.json()
} } }))

const wrappers: VueWrapper[] = []
let fetchMock: ReturnType<typeof vi.fn>
const key = (id = 1, platform = 'openai') => ({ id, key: `synthetic-key-${id}`, group: { id, platform } }) as ApiKey
const response = (payload: unknown, status = 200) => ({ ok: status >= 200 && status < 300, status, json: async () => payload })

function createWrapper(show = true, apiKey = key()) {
  const wrapper = mount(CcSwitchImportModal, {
    props: { show, apiKey, publicSettings: null },
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
        Select: { name: 'Select', props: ['modelValue', 'options', 'loading'], template: '<div class="model-select" />' }
      }
    }
  })
  wrappers.push(wrapper)
  return wrapper
}

function mainSelect(wrapper: VueWrapper) { return wrapper.findAllComponents({ name: 'Select' })[0] }

beforeEach(() => {
  fetchMock = vi.fn()
  vi.stubGlobal('fetch', vi.fn(() => { throw new Error('Must not send keys through a cross-origin redirect') }))
})

afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('CC Switch import model discovery', () => {
  it('loads the selected key through the panel endpoint without a raw key fetch', async () => {
    fetchMock.mockResolvedValue(response({ data: [{ id: 'gpt-6.1-sol' }, { id: 'gpt-6-astra' }] }))
    const wrapper = createWrapper()
    await flushPromises()
    expect(fetchMock).toHaveBeenCalledWith(1, { signal: expect.any(AbortSignal) })
    expect(fetch).not.toHaveBeenCalled()
    expect(JSON.stringify(fetchMock.mock.calls)).not.toContain('synthetic-key')
    expect(mainSelect(wrapper).props('options')).toEqual([{ value: 'gpt-6.1-sol', label: 'gpt-6.1-sol' }, { value: 'gpt-6-astra', label: 'gpt-6-astra' }])
    expect(mainSelect(wrapper).props('modelValue')).toBe('gpt-6.1-sol')
    expect(mainSelect(wrapper).props('loading')).toBe(false)
  })

  it('loads when a closed modal opens and accepts Codex catalogues', async () => {
    fetchMock.mockResolvedValue(response({ models: [{ slug: 'gpt-6-astra' }] }))
    const wrapper = createWrapper(false)
    expect(fetchMock).not.toHaveBeenCalled()
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(mainSelect(wrapper).props('options')).toEqual([{ value: 'gpt-6-astra', label: 'gpt-6-astra' }])
  })

  it('reports empty model metadata instead of inventing IDs or making a cross-origin request', async () => {
    fetchMock.mockResolvedValue(response({ data: [] }))
    const wrapper = createWrapper()
    await flushPromises()
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetch).not.toHaveBeenCalled()
    expect(mainSelect(wrapper).props('options')).toHaveLength(0)
    expect(wrapper.get('[role="alert"]').text()).toContain('keys.ccsImport.modelsEmpty')
  })

  it('shows an actionable failure and supports retry without disclosing response contents', async () => {
    fetchMock.mockResolvedValue(response({ error: { message: 'must not display upstream secrets' } }, 403))
    const wrapper = createWrapper()
    await flushPromises()
    expect(wrapper.get('[data-testid="ccs-models-error"]').text()).toContain('HTTP 403')
    expect(wrapper.text()).not.toContain('must not display upstream secrets')
    fetchMock.mockResolvedValue(response({ data: [{ id: 'gpt-6-astra' }] }))
    await wrapper.get('[data-testid="ccs-models-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(mainSelect(wrapper).props('options')).toHaveLength(1)
  })

  it('ignores stale results after the selected key changes', async () => {
    let resolveOld!: (value: ReturnType<typeof response>) => void
    fetchMock.mockReturnValueOnce(new Promise(resolve => { resolveOld = resolve }))
      .mockResolvedValue(response({ data: [{ id: 'new-key-model' }] }))
    const wrapper = createWrapper()
    const oldSignal = fetchMock.mock.calls[0][1].signal as AbortSignal
    await wrapper.setProps({ apiKey: key(2) })
    await flushPromises()
    expect(oldSignal.aborted).toBe(true)
    resolveOld(response({ data: [{ id: 'old-key-model' }] }))
    await flushPromises()
    expect(mainSelect(wrapper).props('options')).toEqual([{ value: 'new-key-model', label: 'new-key-model' }])
  })

  it('keeps a model manually entered while loading', async () => {
    let resolve!: (value: ReturnType<typeof response>) => void
    fetchMock.mockReturnValue(new Promise(done => { resolve = done }))
    const wrapper = createWrapper()
    mainSelect(wrapper).vm.$emit('update:modelValue', 'manually-entered-model')
    resolve(response({ data: [{ id: 'available-model' }] }))
    await flushPromises()
    expect(mainSelect(wrapper).props('modelValue')).toBe('manually-entered-model')
  })

  it('times out a stalled panel lookup and allows a clean retry', async () => {
    vi.useFakeTimers()
    fetchMock.mockImplementationOnce((_url, options: RequestInit) => new Promise((_resolve, reject) => {
      options.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    })).mockResolvedValue(response({ data: [{ id: 'fallback-model' }] }))
    const wrapper = createWrapper()
    await vi.advanceTimersByTimeAsync(8000)
    await flushPromises()
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(mainSelect(wrapper).props('loading')).toBe(false)
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    await wrapper.get('[data-testid="ccs-models-retry"]').trigger('click')
    await flushPromises()
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(mainSelect(wrapper).props('options')).toHaveLength(1)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it('cancels pending requests when the modal closes', async () => {
    fetchMock.mockImplementation((_url, options: RequestInit) => new Promise((_resolve, reject) => {
      options.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    }))
    const wrapper = createWrapper()
    const signal = fetchMock.mock.calls[0][1].signal as AbortSignal
    await wrapper.setProps({ show: false })
    await flushPromises()
    expect(signal.aborted).toBe(true)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(mainSelect(wrapper).props('loading')).toBe(false)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })
})
