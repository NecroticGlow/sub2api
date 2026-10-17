<template>
  <BaseDialog :show="!!account" :title="t('admin.accounts.sessionExport.title')" width="normal" @close="close">
    <form class="space-y-4" @submit.prevent="exportSession">
      <p class="text-sm font-medium">{{ account?.name }}</p>
      <p v-if="loading" role="status" class="text-sm">{{ t('common.loading') }}</p>
      <template v-else>
        <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('admin.accounts.sessionExport.hint') }}</p>
        <p v-if="configured" class="text-sm text-emerald-700 dark:text-emerald-300">{{ t('admin.accounts.sessionExport.configured') }}</p>
        <label class="block text-sm">
          {{ t(configured ? 'admin.accounts.sessionExport.replaceCookie' : 'admin.accounts.sessionExport.cookie') }}
          <input v-model="cookie" type="password" autocomplete="off" class="input mt-2 w-full font-mono" :disabled="busy" :placeholder="t('admin.accounts.sessionExport.placeholder')" />
        </label>
        <label v-if="storageEnabled" class="flex items-center gap-2 text-sm">
          <input v-model="remember" type="checkbox" :disabled="busy" />
          {{ t('admin.accounts.sessionExport.remember') }}
        </label>
        <p v-else class="text-sm text-amber-700 dark:text-amber-300">{{ t('admin.accounts.sessionExport.storageDisabled') }}</p>
        <p class="text-sm text-amber-700 dark:text-amber-300">{{ t('admin.accounts.sessionExport.sensitive') }}</p>
        <button v-if="configured" type="button" data-testid="delete-cookie" class="text-sm text-red-600" :disabled="busy" @click="deleteCookie">{{ t('admin.accounts.sessionExport.deleteCookie') }}</button>
      </template>
      <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex justify-end gap-2">
        <button type="button" data-testid="close" class="btn btn-secondary" @click="close">{{ t('common.cancel') }}</button>
        <button type="submit" class="btn btn-primary" :disabled="loading || busy || (!configured && !cookie.trim())">{{ t(busy ? 'admin.accounts.sessionExport.exporting' : 'admin.accounts.sessionExport.download') }}</button>
      </div>
    </form>
  </BaseDialog>
  <TotpStepUpDialog :controller="stepUp" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'
import { useStepUp, StepUpCancelledError, isStepUpCancelled, isStepUpBlocked, stepUpBlockReason } from '@/composables/useStepUp'
import { exportOpenAISession, getOpenAISessionCookie, deleteOpenAISessionCookie } from '@/api/admin/accounts'
import type { Account } from '@/types'

const props = withDefaults(defineProps<{ account: Account | null; autoExport?: boolean }>(), { autoExport: true })
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const cookie = ref(''), error = ref(''), busy = ref(false)
const configured = ref(false), storageEnabled = ref(false), remember = ref(false), loading = ref(false)
const stepUp = useStepUp()
let generation = 0

function clear() {
  generation++
  cookie.value = ''; error.value = ''; busy.value = false; loading.value = false
  configured.value = false; storageEnabled.value = false; remember.value = false
  stepUp.onCancel()
}
function close() { clear(); emit('close') }
onBeforeUnmount(clear)

function showError(err: unknown) {
  if (isStepUpCancelled(err)) return
  const reason = (err as { reason?: string })?.reason
  const errorKey = reason?.startsWith('OPENAI_SESSION_') ? reason : 'failed'
  error.value = isStepUpBlocked(err)
    ? t(stepUpBlockReason(err) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' ? 'stepUp.adminApiKeyForbidden' : 'stepUp.notEnabled')
    : t('admin.accounts.sessionExport.' + errorKey)
}

// Drop late errors before useStepUp can open a verification prompt.
function runCurrent<T>(current: number, action: () => Promise<T>): Promise<T> {
  return stepUp.run(async () => {
    if (current !== generation) throw new StepUpCancelledError()
    try { return await action() }
    catch (err) {
      if (current !== generation) throw new StepUpCancelledError()
      throw err
    }
  })
}

async function exportSession() {
  if (!props.account || busy.value || loading.value || (!configured.value && !cookie.value.trim())) return
  const id = props.account.id, current = generation, loginCookie = cookie.value.trim()
  const saveCookie = !!loginCookie && remember.value && storageEnabled.value
  busy.value = true; error.value = ''
  try {
    const session = await runCurrent(current, () => exportOpenAISession(id, loginCookie, saveCookie))
    if (current !== generation) return
    const url = URL.createObjectURL(new Blob([JSON.stringify(session, null, 2)], { type: 'application/json' }))
    const link = document.createElement('a')
    link.href = url; link.download = `chatgpt-session-${id}.json`
    document.body.appendChild(link); link.click(); link.remove()
    setTimeout(() => URL.revokeObjectURL(url), 0)
    close()
  } catch (err) {
    if (current === generation) showError(err)
  } finally {
    if (current === generation) { busy.value = false; cookie.value = '' }
  }
}

async function deleteCookie() {
  if (!props.account || busy.value || loading.value) return
  const id = props.account.id, current = generation
  busy.value = true; error.value = ''
  try {
    await runCurrent(current, () => deleteOpenAISessionCookie(id))
    if (current === generation) { configured.value = false; cookie.value = '' }
  } catch (err) {
    if (current === generation) showError(err)
  } finally {
    if (current === generation) busy.value = false
  }
}

watch(() => props.account, async (account) => {
  clear()
  if (!account) return
  const current = generation
  loading.value = true
  try {
    const state = await getOpenAISessionCookie(account.id)
    if (current !== generation) return
    configured.value = state.configured
    storageEnabled.value = state.storage_enabled
    remember.value = state.storage_enabled
    loading.value = false
    if (state.configured && props.autoExport) await exportSession()
  } catch (err) {
    if (current === generation) showError(err)
  } finally {
    if (current === generation) loading.value = false
  }
}, { immediate: true })
</script>
