<template>
  <div class="space-y-0.5 text-[10px]" data-test="claude-reset-cell">
    <div class="flex items-center gap-1.5">
      <button type="button" :disabled="busy" data-test="query" class="text-blue-600 disabled:opacity-50" @click="refresh">{{ t('admin.accounts.claudeQuotaReset.query') }}</button>
      <span>{{ t('admin.accounts.claudeQuotaReset.count') }} {{ currentCount ?? '—' }}</span>
      <button type="button" :disabled="busy || !selected" data-test="reset" class="text-amber-600 disabled:opacity-40" @click="openConfirm">{{ t('admin.accounts.claudeQuotaReset.use') }}</button>
    </div>
    <div v-for="grant in currentGrants" :key="grant.id" class="text-gray-500">
      {{ grant.resets_left }} × {{ grant.ends_at ? t('admin.accounts.claudeQuotaReset.expires', { time: new Date(grant.ends_at).toLocaleString() }) : t('admin.accounts.claudeQuotaReset.noExpiry') }}
    </div>
    <div v-if="status && !status.eligible" class="text-amber-600">{{ t('admin.accounts.claudeQuotaReset.ineligible') }}{{ status.ineligible_reason ? ` (${status.ineligible_reason})` : '' }}</div>
    <div v-if="status" class="text-gray-400">{{ t('admin.accounts.claudeQuotaReset.updated', { time: new Date(status.fetched_at).toLocaleString() }) }}</div>
    <div v-if="message" role="status" class="max-w-[240px] text-amber-600">{{ message }}</div>
    <ConfirmDialog :show="confirmOpen" :title="t('admin.accounts.claudeQuotaReset.use')" :message="t('admin.accounts.claudeQuotaReset.confirm')"
      :confirm-text="t('admin.accounts.claudeQuotaReset.use')" :cancel-text="t('common.cancel')" danger @confirm="consume" @cancel="confirmOpen = false" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import { refreshClaudeResetCredits, resetClaudeQuota, type ClaudeResetInput, type ClaudeResetStatus } from '@/api/admin/accounts'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const props = defineProps<{ account: Account; credits?: ClaudeResetStatus | null }>()
const emit = defineEmits<{ 'account-updated': [account: Account]; 'reset-completed': [] }>()
const { t } = useI18n()
const status = ref<ClaudeResetStatus | null>(null)
const busy = ref(false)
const message = ref('')
const confirmOpen = ref(false)
const confirmedInput = ref<ClaudeResetInput | null>(null)
let generation = 0
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => { timer = setInterval(() => { now.value = Date.now() }, 30_000) })
onBeforeUnmount(() => { if (timer) clearInterval(timer) })
watch(() => [props.account.id, props.account.extra?.claude_reset_credit_snapshot, props.credits] as const, (value, previous) => {
  generation++
  status.value = props.credits !== undefined ? props.credits : (props.account.extra?.claude_reset_credit_snapshot as ClaudeResetStatus | undefined) ?? null
  confirmOpen.value = false
  confirmedInput.value = null
  if (value[0] !== previous?.[0]) message.value = ''
}, { immediate: true })
const currentGrants = computed(() => status.value?.grants?.filter(g => g.resets_left > 0 && (!g.ends_at || Date.parse(g.ends_at) > now.value)) ?? [])
const currentCount = computed(() => {
  if (!status.value || !Number.isFinite(Date.parse(status.value.fetched_at)) || now.value - Date.parse(status.value.fetched_at) > 24 * 3600_000) return null
  return currentGrants.value.reduce((sum, g) => sum + g.resets_left, 0)
})
const selected = computed(() => {
  const s = status.value
  if (!s?.eligible || now.value - Date.parse(s.fetched_at) > 24 * 3600_000 || !Number.isFinite(Date.parse(s.fetched_at))) return null
  if (s.cooldown_until && !(Date.parse(s.cooldown_until) <= now.value)) return null
  return currentGrants.value.find(g => g.id === s.next_grant_id && g.usable_now && !g.paused && (!g.starts_at || Date.parse(g.starts_at) <= now.value)) ?? null
})
function openConfirm() {
  if (!selected.value) return
  confirmedInput.value = { confirmed: true, grant_id: selected.value.id, resets_left: selected.value.resets_left }
  confirmOpen.value = true
}
async function refresh() {
  if (busy.value) return
  const id = props.account.id, gen = generation
  busy.value = true
  message.value = ''
  try {
    const data = await refreshClaudeResetCredits(id)
    if (id !== props.account.id || gen !== generation) return
    status.value = data
    if (!data) message.value = t('admin.accounts.claudeQuotaReset.unknown')
  } catch {
    if (id === props.account.id && gen === generation) message.value = t('admin.accounts.claudeQuotaReset.queryFailed')
  } finally { busy.value = false }
}
async function consume() {
  confirmOpen.value = false
  if (busy.value || !confirmedInput.value) return
  const id = props.account.id, gen = generation, input = { ...confirmedInput.value }
  busy.value = true
  message.value = ''
  try {
    const data = await resetClaudeQuota(id, input)
    if (id !== props.account.id || gen !== generation) return
    status.value = data.credits
    message.value = t(`admin.accounts.claudeQuotaReset.results.${data.result}`)
    if (data.refresh_warning) message.value += ` ${t('admin.accounts.claudeQuotaReset.refreshWarning')}`
    if (data.account) emit('account-updated', data.account)
    if (data.result === 'reset' || data.result === 'already_used') emit('reset-completed')
  } catch {
    if (id === props.account.id && gen === generation) {
      status.value = null
      message.value = t('admin.accounts.claudeQuotaReset.outcomeUnknown')
    }
  } finally { busy.value = false; confirmedInput.value = null }
}
</script>
