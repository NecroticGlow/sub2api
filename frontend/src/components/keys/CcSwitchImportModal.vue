<template>
  <BaseDialog :show="show" :title="t('keys.ccsImport.title')" width="narrow" @close="emit('close')">
    <div class="space-y-4">
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.app') }}
        </label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="option in appOptions"
            :key="option.value"
            type="button"
            @click="app = option.value"
            :class="[
              'rounded-lg border px-3.5 py-1.5 text-sm font-medium transition-colors',
              app === option.value
                ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300'
                : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300'
            ]"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.name') }}
        </label>
        <input v-model="providerName" type="text" class="input" :placeholder="t('keys.ccsImport.namePlaceholder')" />
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.mainModel') }} <span class="text-red-500">*</span>
          <span v-if="loadingModels" class="ml-2 text-xs font-normal text-gray-400">
            {{ t('keys.ccsImport.loadingModels') }}
          </span>
        </label>
        <Select v-model="mainModel" :options="modelOptions" searchable creatable :loading="loadingModels"
          :placeholder="t('keys.ccsImport.modelPlaceholder')" />
        <div v-if="modelsError" class="mt-2 text-xs text-red-500" role="alert" data-testid="ccs-models-error">
          {{ modelsError }}
          <button type="button" class="ml-2 underline" :disabled="loadingModels" data-testid="ccs-models-retry" @click="fetchModels">
            {{ t('keys.ccsImport.retryModels') }}
          </button>
        </div>
      </div>

      <template v-if="app === 'claude'">
        <div v-for="field in tieredFields" :key="field.key">
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t(`keys.ccsImport.${field.key}Model`) }}
          </label>
          <Select v-model="field.value.value" :options="modelOptions" searchable creatable
            :placeholder="t('keys.ccsImport.modelPlaceholder')" />
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button @click="emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button @click="handleOpen" :disabled="!mainModel.trim()" class="btn btn-primary">
          {{ t('keys.ccsImport.open') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { useAppStore } from '@/stores'
import type { ApiKey, PublicSettings } from '@/types'
import {
  CC_SWITCH_PROVIDER_API_BASE_URL,
  buildCcSwitchImportDeeplink,
  ccSwitchModelsUrls,
  parseCcSwitchModelIds,
  defaultCcSwitchAppForPlatform,
  defaultCcSwitchModelForPlatform,
  type CcSwitchApp
} from '@/utils/ccswitchImport'

const props = defineProps<{ show: boolean; apiKey: ApiKey | null; publicSettings: PublicSettings | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const appStore = useAppStore()

const app = ref<CcSwitchApp>('claude')
const providerName = ref('')
const mainModel = ref('')
const haikuModel = ref('')
const sonnetModel = ref('')
const opusModel = ref('')
const models = ref<string[]>([])
const loadingModels = ref(false)
const modelsError = ref('')
let modelsRequestId = 0
let modelsController: AbortController | null = null
const platform = computed(() => props.apiKey?.group?.platform || 'anthropic')
const gatewayBaseUrl = computed(() =>
  (props.publicSettings?.api_base_url || window.location.origin).replace(/\/+$/, '')
)
const modelOptions = computed(() => models.value.map(model => ({ value: model, label: model })))
const tieredFields = computed(() => [
  { key: 'haiku', value: haikuModel },
  { key: 'sonnet', value: sonnetModel },
  { key: 'opus', value: opusModel }
])
const appOptions = computed<{ value: CcSwitchApp; label: string }[]>(() => {
  const options: { value: CcSwitchApp; label: string }[] = [
    { value: 'claude', label: 'Claude' },
    { value: 'codex', label: 'Codex' },
    { value: 'gemini', label: 'Gemini' },
    { value: 'opencode', label: 'OpenCode' }
  ]
  if (platform.value === 'grok') options.push({ value: 'grokbuild', label: t('keys.ccsImport.grokBuild') })
  return options
})

function resetForm() {
  app.value = defaultCcSwitchAppForPlatform(platform.value)
  providerName.value = (props.publicSettings?.site_name || 'sub2api').trim() || 'sub2api'
  mainModel.value = defaultCcSwitchModelForPlatform(platform.value)
  haikuModel.value = ''
  sonnetModel.value = ''
  opusModel.value = ''
  models.value = []
  modelsError.value = ''
}

function pickTieredDefaults() {
  const find = (value: string) => models.value.find(model => model.toLowerCase().includes(value)) || ''
  if (!haikuModel.value) haikuModel.value = find('haiku')
  if (!sonnetModel.value) sonnetModel.value = find('sonnet')
  if (!opusModel.value) opusModel.value = find('opus')
}

function cancelModelsRequest() {
  modelsRequestId++
  modelsController?.abort()
  modelsController = null
  loadingModels.value = false
}

async function fetchModels() {
  cancelModelsRequest()
  const row = props.apiKey
  if (!row) return
  const requestId = modelsRequestId
  const initialModel = mainModel.value
  let detail = t('keys.ccsImport.modelsNetworkError')
  loadingModels.value = true
  modelsError.value = ''
  try {
    for (const url of ccSwitchModelsUrls(gatewayBaseUrl.value, window.location.origin, platform.value)) {
      if (requestId !== modelsRequestId) return
      const controller = new AbortController()
      modelsController = controller
      const timeout = window.setTimeout(() => controller.abort(), 8000)
      try {
        const response = await fetch(url, { headers: { Authorization: `Bearer ${row.key}` }, signal: controller.signal })
        if (requestId !== modelsRequestId) return
        if (!response.ok) {
          detail = `HTTP ${response.status}`
          continue
        }
        const available = parseCcSwitchModelIds(await response.json())
        if (requestId !== modelsRequestId) return
        if (!available.length) {
          detail = t('keys.ccsImport.modelsEmpty')
          continue
        }
        models.value = available
        // Do not overwrite a model manually entered while discovery was pending.
        if (mainModel.value === initialModel && !available.includes(initialModel)) mainModel.value = available[0]
        pickTieredDefaults()
        return
      } catch {
        if (requestId !== modelsRequestId) return
      } finally {
        window.clearTimeout(timeout)
        if (modelsController === controller) modelsController = null
      }
    }
    if (requestId === modelsRequestId) modelsError.value = t('keys.ccsImport.modelsLoadFailed', { detail })
  } finally {
    if (requestId === modelsRequestId) loadingModels.value = false
  }
}

watch([
  () => props.show,
  () => props.apiKey?.id,
  () => props.apiKey?.key,
  () => props.apiKey?.group?.id,
  platform,
  gatewayBaseUrl
], ([show]) => {
  cancelModelsRequest()
  if (show && props.apiKey) {
    resetForm()
    void fetchModels()
  }
}, { immediate: true })

onUnmounted(cancelModelsRequest)

function buildUsageScript(): string {
  return `({
    request: { url: "${CC_SWITCH_PROVIDER_API_BASE_URL}/v1/usage", method: "GET", headers: { "Authorization": "Bearer {{apiKey}}" } },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return { isValid: response?.is_active ?? response?.isValid ?? true, remaining, unit };
    }
  })`
}

function handleOpen() {
  const row = props.apiKey
  if (!row || !mainModel.value.trim()) return
  const deeplink = buildCcSwitchImportDeeplink({
    homepage: window.location.origin, platform: row.group?.platform, app: app.value,
    providerName: providerName.value.trim() || 'sub2api', apiKey: row.key,
    usageScript: buildUsageScript(), model: mainModel.value,
    haikuModel: haikuModel.value, sonnetModel: sonnetModel.value, opusModel: opusModel.value
  })
  try {
    window.open(deeplink, '_self')
    setTimeout(() => {
      if (document.hasFocus()) appStore.showError(t('keys.ccSwitchNotInstalled'))
    }, 100)
  } catch {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
  emit('close')
}
</script>
