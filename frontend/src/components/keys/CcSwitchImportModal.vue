<template>
  <BaseDialog :show="show" :title="t('keys.ccsImport.title')" width="narrow" @close="emit('close')">
    <div class="space-y-4">
      <!-- App selection -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.app') }}
        </label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="option in appOptions"
            :key="option.value"
            type="button"
            @click="selectApp(option.value)"
            :class="[
              'rounded-lg border px-3.5 py-1.5 text-sm font-medium transition-colors',
              app === option.value
                ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300'
                : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:text-gray-800 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300 dark:hover:border-dark-500 dark:hover:text-dark-100'
            ]"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <!-- Provider name -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.name') }}
        </label>
        <input
          v-model="providerName"
          type="text"
          class="input"
          :placeholder="t('keys.ccsImport.namePlaceholder')"
        />
      </div>

      <!-- Main model (required) -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.ccsImport.mainModel') }}
          <span class="text-red-500">*</span>
          <span v-if="loadingModels" class="ml-2 text-xs font-normal text-gray-400">
            {{ t('keys.ccsImport.loadingModels') }}
          </span>
        </label>
        <Select
          v-model="mainModel"
          :options="modelOptions"
          searchable
          creatable
          :loading="loadingModels"
          :placeholder="t('keys.ccsImport.modelPlaceholder')"
        />
      </div>

      <!-- Claude tiered models -->
      <template v-if="app === 'claude'">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('keys.ccsImport.haikuModel') }}
          </label>
          <Select
            v-model="haikuModel"
            :options="modelOptions"
            searchable
            creatable
            :placeholder="t('keys.ccsImport.modelPlaceholder')"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('keys.ccsImport.sonnetModel') }}
          </label>
          <Select
            v-model="sonnetModel"
            :options="modelOptions"
            searchable
            creatable
            :placeholder="t('keys.ccsImport.modelPlaceholder')"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('keys.ccsImport.opusModel') }}
          </label>
          <Select
            v-model="opusModel"
            :options="modelOptions"
            searchable
            creatable
            :placeholder="t('keys.ccsImport.modelPlaceholder')"
          />
        </div>
      </template>

    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button @click="handleOpen" :disabled="!canOpen" class="btn btn-primary">
          {{ t('keys.ccsImport.open') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { useAppStore } from '@/stores'
import type { ApiKey, PublicSettings } from '@/types'
import {
  buildCcSwitchImportDeeplink,
  ccSwitchModelsUrls,
  defaultCcSwitchAppForPlatform,
  defaultCcSwitchModelForPlatform,
  type CcSwitchApp
} from '@/utils/ccswitchImport'

const props = defineProps<{
  show: boolean
  apiKey: ApiKey | null
  publicSettings: PublicSettings | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

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

const platform = computed(() => props.apiKey?.group?.platform || 'anthropic')

const appOptions = computed<{ value: CcSwitchApp; label: string }[]>(() => {
  const options: { value: CcSwitchApp; label: string }[] = [
    { value: 'claude', label: 'Claude' },
    { value: 'codex', label: 'Codex' },
    { value: 'gemini', label: 'Gemini' },
    { value: 'opencode', label: 'OpenCode' }
  ]
  if (platform.value === 'grok') {
    options.push({ value: 'grokbuild', label: t('keys.ccsImport.grokBuild') })
  }
  return options
})

const canOpen = computed(() => mainModel.value.trim().length > 0)
const modelOptions = computed(() => models.value.map((model) => ({ value: model, label: model })))

const gatewayBaseUrl = computed(() =>
  (props.publicSettings?.api_base_url || window.location.origin).replace(/\/+$/, '')
)

function selectApp(value: CcSwitchApp) {
  app.value = value
}

function resetForm() {
  app.value = defaultCcSwitchAppForPlatform(platform.value)
  providerName.value = (props.publicSettings?.site_name || 'sub2api').trim() || 'sub2api'
  mainModel.value = defaultCcSwitchModelForPlatform(platform.value)
  haikuModel.value = ''
  sonnetModel.value = ''
  opusModel.value = ''
  models.value = []
}

function pickTieredDefaults() {
  const lower = (id: string) => id.toLowerCase()
  if (!haikuModel.value) {
    haikuModel.value = models.value.find((m) => lower(m).includes('haiku')) || ''
  }
  if (!sonnetModel.value) {
    sonnetModel.value = models.value.find((m) => lower(m).includes('sonnet')) || ''
  }
  if (!opusModel.value) {
    opusModel.value = models.value.find((m) => lower(m).includes('opus')) || ''
  }
}

async function fetchModels() {
  const row = props.apiKey
  if (!row) return
  loadingModels.value = true
  try {
    for (const url of ccSwitchModelsUrls(gatewayBaseUrl.value, window.location.origin)) {
      try {
        const response = await fetch(url, {
          headers: { Authorization: `Bearer ${row.key}` }
        })
        if (!response.ok) continue
        const payload: unknown = await response.json()
        const data = (payload as { data?: unknown })?.data
        if (!Array.isArray(data)) continue
        const ids = data
          .map((item) => (item as { id?: unknown })?.id)
          .filter((id): id is string => typeof id === 'string' && id.length > 0)
        models.value = Array.from(new Set(ids))
        if (models.value.length > 0) break
      } catch {
        // Try the configured public endpoint after a same-origin failure.
      }
    }
    if (!mainModel.value && models.value.length > 0) {
      mainModel.value = models.value[0]
    }
    pickTieredDefaults()
  } catch {
    // Model list is a convenience; manual input still works.
  } finally {
    loadingModels.value = false
  }
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      resetForm()
      void fetchModels()
    }
  }
)

function buildUsageScript(): string {
  return `({
    request: {
      url: "{{baseUrl}}/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`
}

function handleOpen() {
  const row = props.apiKey
  if (!row || !canOpen.value) return

  const baseUrl = props.publicSettings?.api_base_url || window.location.origin
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl,
    platform: row.group?.platform,
    app: app.value,
    providerName: providerName.value.trim() || 'sub2api',
    apiKey: row.key,
    usageScript: buildUsageScript(),
    model: mainModel.value,
    haikuModel: haikuModel.value,
    sonnetModel: sonnetModel.value,
    opusModel: opusModel.value
  })

  try {
    window.open(deeplink, '_self')
    // If the protocol handler fails, focus stays on the page.
    setTimeout(() => {
      if (document.hasFocus()) {
        appStore.showError(t('keys.ccSwitchNotInstalled'))
      }
    }, 100)
  } catch {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
  emit('close')
}
</script>
