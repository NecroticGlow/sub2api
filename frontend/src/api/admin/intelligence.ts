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
  usage?: {
    input_tokens: number
    output_tokens: number
    total_tokens: number
    cached_input_tokens?: number
  }
  cost?: {
    input_cost_usd: number
    cached_input_cost_usd: number
    output_cost_usd: number
    total_cost_usd: number
    pricing_note: string
  }
  current_concurrency: number
}

export async function testIntelligence(id: number, signal?: AbortSignal): Promise<IntelligenceResult> {
  const { data } = await apiClient.post<IntelligenceResult>(`/admin/accounts/${id}/intelligence-test`, {}, {
    timeout: 165000,
    signal
  })
  return data
}

export async function getIntelligenceHistory(id: number, signal?: AbortSignal): Promise<IntelligenceResult[]> {
  const { data } = await apiClient.get<{ items: IntelligenceResult[] }>(`/admin/accounts/${id}/intelligence-test-history`, {
    timeout: 15000,
    signal
  })
  return data.items ?? []
}
