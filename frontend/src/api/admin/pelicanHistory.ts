import { apiClient } from '../client'
import type { IntelligenceQuestion } from '@/utils/intelligenceTest'

export interface IntelligenceRun {
  questionKind?: IntelligenceQuestion
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
  questionKind?: IntelligenceQuestion
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
