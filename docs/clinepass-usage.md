# DeepSeek ClinePass 用量查询

DeepSeek 平台的 API Key 账号，`base_url` 主机为 `https://api.cline.bot`
（含 `/api/v1`、`/v1` 等路径），会自动使用 ClinePass 订阅用量查询。
不需要修改现有按量/Coding 模式，也不需要重新录入账号。

新增或编辑 DeepSeek 账号时，可使用 **ClinePass 用量查询** 开关。
设置持久化在账号 credentials 的 `clinepass_usage_enabled` 布尔字段，
所有管理员共享；未设置的旧账号保持自动识别。新账号默认开启，
但仍只对上述 Cline 官方地址生效，不按账号 ID 或账号名识别。
关闭后停止手动查询、后台定时刷新和渠道监控额度探测，保留已有快照及
用量日志，不影响模型转发，也不会回落到 DeepSeek 官方余额接口。

查询固定请求 `GET https://api.cline.bot/api/v1/users/me/plan/usage-limits`，
请求头为 `Authorization: Bearer <账号 api_key>`。请求遵守服务器出站白名单及
账号代理设置；启用出站白名单的部署需要允许 `api.cline.bot`。
未识别为 Cline 官方账号的中转 Key 不会发送到这个地址。

账号列表用量列显示 ClinePass 的 5 小时、每周、每月已用百分比和重置时间。
支持手动查询，并复用现有过期快照刷新和后端周期查询机制。
结果持久化在账号 Extra 的 `deepseek_5h_*`、`deepseek_weekly_*`、
`deepseek_monthly_*` 字段，其他管理员可以直接查看。查询失败不覆盖有效快照。
空响应、无订阅或无有效窗口显示错误，不当成 0% 用量。

DeepSeek 缓存估算已移除，包括 LRU 前缀指纹、概率估算、各协议响应注入以及
旧设置。缓存计费只依据上游真实报告的命中数量；缺失时不虚构缓存命中。
已有用量日志和 DeepSeek 峰谷价格不变。
