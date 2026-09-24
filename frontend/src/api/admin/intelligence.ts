import { apiClient } from '../client'

export interface IntelligenceResult {
  account_id: number
  model: string
  status: 'passed' | 'manual_review' | 'error'
  response_text: string
  error?: string
  checks: { item: string; expected: string; matched: boolean }[]
  tested_at: string
  latency_ms: number
}

export async function testIntelligence(id: number, signal?: AbortSignal): Promise<IntelligenceResult> {
  const { data } = await apiClient.post<IntelligenceResult>(`/admin/accounts/${id}/intelligence-test`, {}, {
    timeout: 165000,
    signal
  })
  return data
}
