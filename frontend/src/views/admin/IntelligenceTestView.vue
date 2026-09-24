<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="card p-5">
        <h1 class="text-xl font-semibold">{{ t('admin.intelligence.title') }}</h1>
        <p class="mt-2 text-sm text-gray-500">{{ t('admin.intelligence.description') }}</p>
        <p class="mt-2 text-sm text-gray-500">{{ t('admin.intelligence.note') }}</p>
        <div class="mt-4 flex flex-wrap items-center gap-3">
          <span class="rounded bg-primary-50 px-3 py-1 font-mono text-sm text-primary-700 dark:bg-dark-700 dark:text-primary-300">gpt-6-astra</span>
          <form class="flex flex-1 gap-2" @submit.prevent="page = 1; loadAccounts()">
            <input v-model="search" class="input min-w-0 max-w-sm" :aria-label="t('admin.intelligence.search')" :placeholder="t('admin.intelligence.search')" />
            <button class="btn btn-secondary" :disabled="loading">{{ t('common.search') }}</button>
          </form>
          <button class="btn btn-secondary" :disabled="loading" @click="loadAccounts">{{ t('common.refresh') }}</button>
        </div>
      </div>
      <p v-if="loadError" role="alert" class="text-red-600">{{ loadError }}</p>
      <p v-if="loading" role="status">{{ t('common.loading') }}</p>
      <p v-else-if="!accounts.length && !loadError" class="card p-8 text-center text-gray-500">{{ t('admin.intelligence.empty') }}</p>
      <article v-for="account in accounts" :key="account.id" class="card p-5" :data-account-id="account.id">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="min-w-0">
            <h2 class="break-words font-semibold">{{ account.name }} <span class="text-sm font-normal text-gray-500">#{{ account.id }}</span></h2>
            <p class="mt-1 text-xs text-gray-500">OpenAI · OAuth · {{ account.status }}</p>
          </div>
          <div class="flex items-center gap-3">
            <span role="status" :class="statusClass(account.id)">{{ statusLabel(account.id) }}</span>
            <button class="btn btn-primary" :disabled="!!running[account.id]" @click="runTest(account.id)">
              {{ t(running[account.id] ? 'admin.intelligence.testing' : 'admin.intelligence.test') }}
            </button>
          </div>
        </div>
        <template v-if="results[account.id]">
          <p class="mt-3 text-xs text-gray-500">{{ new Date(results[account.id]!.tested_at).toLocaleString() }} · {{ (results[account.id]!.latency_ms / 1000).toFixed(1) }}s</p>
          <div class="mt-3 grid gap-2 text-sm sm:grid-cols-2 lg:grid-cols-4">
            <span>{{ t('admin.intelligence.concurrency') }}: {{ results[account.id]!.current_concurrency }}</span>
            <span>{{ t('admin.intelligence.tokens') }}: {{ formatNumber(results[account.id]!.usage?.total_tokens) }}</span>
            <span>{{ t('admin.intelligence.inputOutput') }}: {{ formatNumber(results[account.id]!.usage?.input_tokens) }} / {{ formatNumber(results[account.id]!.usage?.output_tokens) }}</span>
            <span>{{ t('admin.intelligence.originalCost') }}: ${{ (results[account.id]!.cost?.total_cost_usd ?? 0).toFixed(6) }}</span>
          </div>
          <p v-if="results[account.id]!.cost" class="mt-1 text-xs text-gray-500">{{ results[account.id]!.cost!.pricing_note }}</p>
          <p v-if="previousResults[account.id]" class="mt-2 text-xs text-gray-500">{{ t('admin.intelligence.previous') }}: {{ previousLabel(account.id) }}</p>
          <p v-if="results[account.id]!.error" role="alert" class="mt-3 whitespace-pre-wrap break-words text-sm text-red-600">{{ results[account.id]!.error }}</p>
          <details class="mt-3" :open="results[account.id]!.status === 'manual_review'">
            <summary class="cursor-pointer text-sm font-medium">{{ t('admin.intelligence.answer') }}</summary>
            <pre class="mt-3 max-h-96 overflow-auto whitespace-pre-wrap break-words rounded-lg bg-gray-50 p-4 text-sm dark:bg-dark-900">{{ results[account.id]!.response_text || '—' }}</pre>
            <ul class="mt-3 space-y-1 text-sm">
              <li v-for="check in results[account.id]!.checks" :key="check.item" :class="check.matched ? 'text-green-600' : 'text-amber-600'">
                {{ check.matched ? '✓' : '○' }} {{ t(`admin.intelligence.items.${check.item}`) }}: {{ check.expected }}
              </li>
            </ul>
          </details>
          <div v-if="results[account.id]!.status === 'manual_review'" class="mt-4 flex flex-wrap items-center gap-2">
            <span class="text-sm">{{ t('admin.intelligence.review') }}</span>
            <button class="btn btn-secondary btn-sm" :aria-pressed="manual[account.id] === 'normal'" @click="manual[account.id] = 'normal'">{{ t('admin.intelligence.normal') }}</button>
            <button class="btn btn-secondary btn-sm" :aria-pressed="manual[account.id] === 'degraded'" @click="manual[account.id] = 'degraded'">{{ t('admin.intelligence.degraded') }}</button>
            <button v-if="manual[account.id]" class="btn btn-secondary btn-sm" @click="delete manual[account.id]">{{ t('admin.intelligence.reset') }}</button>
          </div>
        </template>
      </article>
      <div class="flex items-center justify-between gap-3">
        <button class="btn btn-secondary" :disabled="loading || page <= 1" @click="page--; loadAccounts()">{{ t('pagination.previous') }}</button>
        <span class="text-sm text-gray-500">{{ page }} / {{ pages }} · {{ total }}</span>
        <button class="btn btn-secondary" :disabled="loading || page >= pages" @click="page++; loadAccounts()">{{ t('pagination.next') }}</button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { list } from '@/api/admin/accounts'
import { testIntelligence, type IntelligenceResult } from '@/api/admin/intelligence'
import type { AccountListItem } from '@/types'

const { t } = useI18n()
const accounts = ref<AccountListItem[]>([])
const results = ref<Record<number, IntelligenceResult>>({})
const previousResults = ref<Record<number, IntelligenceResult>>({})
const previousManual = ref<Record<number, 'normal' | 'degraded'>>({})
const manual = ref<Record<number, 'normal' | 'degraded'>>({})
const running = ref<Record<number, boolean>>({})
const controllers = new Map<number, AbortController>()
const loading = ref(false)
const loadError = ref('')
const search = ref('')
const page = ref(1)
const pages = ref(1)
const total = ref(0)
let listController: AbortController | undefined

async function loadAccounts() {
  listController?.abort()
  const controller = new AbortController()
  listController = controller
  loading.value = true
  loadError.value = ''
  try {
    const data = await list(page.value, 20, { platform: 'openai', type: 'oauth', lite: 'true', search: search.value.trim() }, { signal: controller.signal })
    if (controller.signal.aborted) return
    accounts.value = data.items.filter(account => account.platform === 'openai' && account.type === 'oauth')
    total.value = data.total
    pages.value = Math.max(1, data.pages)
  } catch (error) {
    if (!controller.signal.aborted) loadError.value = error instanceof Error ? error.message : t('common.unknownError')
  } finally {
    if (listController === controller) loading.value = false
  }
}

async function runTest(id: number) {
  if (running.value[id]) return
  const controller = new AbortController()
  controllers.set(id, controller)
  running.value[id] = true
  if (results.value[id]) {
    previousResults.value[id] = results.value[id]
    if (manual.value[id]) previousManual.value[id] = manual.value[id]
  }
  delete manual.value[id]
  const started = Date.now()
  try {
    const result = await testIntelligence(id, controller.signal)
    if (!controller.signal.aborted) results.value[id] = result
  } catch (error) {
    if (!controller.signal.aborted) results.value[id] = {
      account_id: id, model: 'gpt-6-astra', status: 'error', response_text: '', checks: [], current_concurrency: 0,
      error: error instanceof Error ? error.message : t('common.unknownError'),
      tested_at: new Date(started).toISOString(), latency_ms: Date.now() - started
    }
  } finally {
    running.value[id] = false
    controllers.delete(id)
  }
}

function statusLabel(id: number) {
  if (running.value[id]) return t('admin.intelligence.testing')
  if (manual.value[id]) return t(`admin.intelligence.reviewed_${manual.value[id]}`)
  return statusLabelFor(results.value[id]?.status ?? 'idle')
}
function statusLabelFor(status: IntelligenceResult['status'] | 'idle') {
  return t(`admin.intelligence.${status}`)
}
function previousLabel(id: number) {
  const decision = previousManual.value[id]
  return decision ? t(`admin.intelligence.reviewed_${decision}`) : statusLabelFor(previousResults.value[id]!.status)
}
function formatNumber(value: number | undefined) {
  return new Intl.NumberFormat().format(value ?? 0)
}
function statusClass(id: number) {
  if (manual.value[id] === 'degraded' || results.value[id]?.status === 'error') return 'text-sm text-red-600'
  if (manual.value[id] === 'normal' || results.value[id]?.status === 'passed') return 'text-sm text-green-600'
  return 'text-sm text-amber-600'
}
onMounted(loadAccounts)
onBeforeUnmount(() => {
  listController?.abort()
  controllers.forEach(controller => controller.abort())
})
</script>
