import type { ScheduledTestResult } from '@/types'

export type IntelligenceQuestion = 'candy' | 'pelican' | 'knowledge' | 'japan_pm'
export const CANDY_PROMPT = `在一个黑色的袋子里放有三种口味的糖果，每种糖果有两种不同的形状（圆形和五角星形，不同的形状靠手感可以分辨）。现已知不同口味的糖和不同形状的数量统计如下表。参赛者需要在活动前决定摸出的糖果数目，那么，最少取出多少个糖果才能保证手中同时拥有不同形状的苹果味和桃子味的糖？（同时手中有圆形苹果味匹配五角星桃子味糖果，或者有圆形桃子味匹配五角星苹果味糖果都满足要求）
苹果味 桃子味 西瓜味
圆形 7 9 8
五角星形 7 6 4`
export const PELICAN_PROMPT = '创建一个 HTML，内容是 SVG 绘制一个鹈鹕骑自行车的 2D 动画，你不需要任何测试'
export const KNOWLEDGE_PROMPT = `仅依据你已有的知识回答，不联网、不调用工具、不猜测。列出你明确知道的最新一代 iPhone 型号、Apple 官方公布日期和正式开售日期，以及你明确知道的最新 NVIDIA GPU 型号、Android 版本、macOS 版本和 Windows 小版本。
“最新”指你已有知识中明确知道的最新版本，不要求确认截至今天是否最新。任何项目或日期不确定，只在对应位置写“uncertain”。直接输出答案，不添加前言、免责声明、解释或后续建议。`
export const JAPAN_PM_PROMPT = '仅依据你已有的知识，直接给出你明确知道的最近一任日本首相姓名。不联网、不调用工具、不猜测，不要求确认其截至今天是否仍在任。如果姓名不确定，只输出“uncertain”。不添加任何解释或免责声明。'
// Maintainer-provided test references, not a live claim about today's latest versions.
export const KNOWLEDGE_REFERENCE = 'iPhone: iPhone 17 series; Apple 官方公布日期: September 9, 2025; 正式开售日期: September 19, 2025; NVIDIA GPU: GeForce RTX 5090; Android: Android 16; macOS: macOS Tahoe 26; Windows minor version: Windows 11, version 25H2'
export const JAPAN_PM_REFERENCE = '高市早苗'
export function questionReferenceAnswer(kind: IntelligenceQuestion): string {
  return kind === 'knowledge' ? KNOWLEDGE_REFERENCE : kind === 'japan_pm' ? JAPAN_PM_REFERENCE : kind === 'candy' ? '21' : ''
}
export function questionPrompt(kind: IntelligenceQuestion): string { return kind === 'candy' ? CANDY_PROMPT : kind === 'knowledge' ? KNOWLEDGE_PROMPT : kind === 'japan_pm' ? JAPAN_PM_PROMPT : PELICAN_PROMPT }
export function questionContract(kind: IntelligenceQuestion): string {
  if (kind === 'knowledge' || kind === 'japan_pm') return '仅依据已有知识，不联网、不调用工具、不猜测。直接输出答案；不确定时按题目要求输出“uncertain”，不添加前言、免责声明、解释或后续建议。'
  return kind === 'candy' ? '只输出最终整数，不要解释。' : '所有账号使用相同交付约定：直接返回独立 HTML，不使用 Markdown 代码块或外部依赖。只输出 HTML，不要解释。'
}

// 定时测试 / 质量规则里的探针题型：不下发题目，后端跑一次门票探针判满血/降智。
export const STATE_PROBE_QUESTION = 'state_probe'
export const DEFAULT_STATE_PROBE_CRON = '*/2 * * * *'
export type StateProbeVerdict = 'healthy' | 'degraded' | 'inconclusive'

// 结果是纯文本（不是 HTML 动画）的题型。
export function isTextAnswerKind(kind?: string): boolean {
  return kind === 'candy' || kind === 'knowledge' || kind === 'japan_pm' || kind === STATE_PROBE_QUESTION
}

// 与后端约定：满血 = success；降智 = failed + state_degraded；其余失败都是无法判断。
export function stateProbeVerdict(result: Pick<ScheduledTestResult, 'status' | 'error_message' | 'pelican_config'>): StateProbeVerdict | null {
  if (result.pelican_config?.question_kind !== STATE_PROBE_QUESTION || result.status === 'running') return null
  if (result.status === 'success') return 'healthy'
  if (result.error_message === 'state_degraded') return 'degraded'
  return 'inconclusive'
}
