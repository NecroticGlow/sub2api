export default {
  intelligence: {
    title: '降智测试',
    description: '仅测试 OpenAI OAuth 账号。使用 gpt-6-astra 和 v2.8.11 降智测试题，不调用联网工具。',
    note: '糖果题返回标准答案 21 时自动标记“未降智”；其他回答待人工判定。鹈鹕题返回包含 SVG 的 HTML 时自动通过。结果保存在服务器，其他管理员可查看。',
    search: '搜索账号名称', empty: '没有 OpenAI OAuth 账号',
    test: '开始测试', testing: '测试中…', idle: '未测试', question: '测试题目', candy: '糖果题（标准答案 21）', pelican: '鹈鹕 SVG HTML', knowledge: '不联网最新信息题', reasoning: '推理强度',
    passed: '未降智（参考答案匹配）', manual_review: '待人工判定', error: '测试失败',
    answer: '查看原始回答与参考项', review: '人工判定：', previous: '上一次测试', lastTest: '上一次测试结果', history: '历史测试情况', concurrency: '当前并发', tokens: '总 token', inputOutput: '输入 / 输出 token', originalCost: '原价额度',
    normal: '未降智', degraded: '降智', reset: '撤销判定',
    reviewed_normal: '未降智（人工判定）', reviewed_degraded: '降智（人工判定）',
    items: { iphone: 'iPhone', announcement: '官方发布日', on_sale: '正式发售日', nvidia: 'NVIDIA GPU', android: 'Android', macos: 'macOS', windows: 'Windows', candy: '糖果题', pelican: '鹈鹕 HTML' }
  }
}
