import { apiClient } from '../client'
import type { StateProbeVerdict } from '@/utils/intelligenceTest'
import type { PelicanTestConfig } from '@/types'

export interface IntelligenceRun {
  questionKind?: PelicanTestConfig['question_kind']
  verdict?: StateProbeVerdict
  id: string
  status: 'running' | 'success' | 'error'
  output: string
  html: string
  error: string
  source?: 'manual' | 'scheduled'
  startedAt?: string
  finishedAt?: string
  durationMs?: number
  modelId?: string
  reasoningEffort?: string
}
export interface IntelligenceRecord {
  questionKind?: PelicanTestConfig['question_kind']
  id: string
  createdAt: string
  prompt: string
  modelId: string
  reasoningEffort: string
  runs: IntelligenceRun[]
}

export const pelicanHistoryAPI = {
  async list(accountId: number): Promise<IntelligenceRecord[]> {
    const { data } = await apiClient.get<IntelligenceRecord[]>(`/admin/accounts/${accountId}/pelican-test-history`)
    return data ?? []
  },
  async save(accountId: number, record: IntelligenceRecord): Promise<void> {
    await apiClient.post(`/admin/accounts/${accountId}/pelican-test-history`, record)
  }
}
