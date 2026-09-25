export default {
  intelligence: {
    title: '降智测试',
    description: '仅测试 OpenAI OAuth 账号。固定使用 gpt-6-astra，根据预设知识问答检查返回内容，不调用联网工具。',
    note: '全部匹配参考答案时标记“未降智”，其他回答待人工判定。此测试仅作筛查，不代表能力鉴定。每次测试消耗上游用量；测试与人工判定结果仅在当前页面保留。',
    search: '搜索账号名称', empty: '没有 OpenAI OAuth 账号',
    test: '开始测试', testing: '测试中…', idle: '未测试',
    passed: '未降智（参考答案匹配）', manual_review: '待人工判定', error: '测试失败',
    answer: '查看原始回答与参考项', review: '人工判定：', previous: '上一次测试', lastTest: '上一次测试结果', concurrency: '当前并发', tokens: '总 token', inputOutput: '输入 / 输出 token', originalCost: '原价额度',
    normal: '未降智', degraded: '降智', reset: '撤销判定',
    reviewed_normal: '未降智（人工判定）', reviewed_degraded: '降智（人工判定）',
    items: { iphone: 'iPhone', announcement: '官方发布日', on_sale: '正式发售日', nvidia: 'NVIDIA GPU', android: 'Android', macos: 'macOS', windows: 'Windows' }
  }
}
