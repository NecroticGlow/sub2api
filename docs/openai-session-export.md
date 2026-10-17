# OpenAI 账号导出 ChatGPT Session

支持持有独立凭据的 OpenAI OAuth 账号；API Key 账号和 Spark 影子账号不支持。

## 首次配置，以后直接下载

1. 在浏览器中登录目标账号的 ChatGPT，访问 `https://chatgpt.com/api/auth/session`，在开发者工具 Network 面板复制该请求的完整 `Cookie` 值。如果使用“Copy as cURL”，复制 `-b` 后引号内的值，不要粘贴整个命令。分块登录 Cookie 必须包括全部 `__Secure-next-auth.session-token.0`、`.1` 等分块。
2. 在账号“更多操作 → Session 登录设置”粘贴 Cookie，勾选“验证成功后加密保存 Cookie”，点击“请求并下载 Session”。服务端请求真实接口，确认 Session 未过期且用户身份匹配后，才保存 Cookie 并下载 `chatgpt-session-<账号ID>.json`。错误或其他用户的 Cookie 不会覆盖旧记录。
3. 以后点击“更多操作 → 导出 ChatGPT Session”，系统自动使用已保存的网页登录凭据请求并下载，无需再次填写 Cookie。敏感操作需要二次验证时，仍需完成验证。
4. Cookie 失效时，界面保留错误提示和更新输入框；可重新粘贴 Cookie。也可以随时通过“Session 登录设置”更新或删除已保存凭据。仅删除本系统保存的副本，不会退出浏览器的 ChatGPT 登录。

首次配置需要服务器设置固定的 `TOTP_ENCRYPTION_KEY`（复用现有 AES-256-GCM 加密能力）。请保留原密钥；更换密钥可能使现有加密数据无法解密。未配置固定密钥时，界面提示仅支持临时导出，不会用重启后失效的临时密钥保存 Cookie。取消保存勾选后，可以只导出本次请求；不会修改已保存记录。

## 登录态与安全边界

cURL 的 `-b` 就是 Cookie 参数。浏览器打开 Session 地址时自动携带自己的 Cookie，因此看起来不需要手动输入；服务端无法自动获得用户浏览器的登录态。这里实现的是“一次配置、后续直接导出”，并不是无登录凭据请求，也不使用 OAuth token 合成 Session。

Cookie 以 AES-GCM 密文保存在独立的 `openai_session_cookies` 表中，不进入普通账号详情、账号 JSON 导出、CRS 同步或审计请求正文。状态接口仅返回是否已配置、是否允许存储，不返回 Cookie 或密文。完整数据库备份包含该密文，恢复时需要原加密密钥；账号软删除会在同一事务中清理 Cookie 记录，物理删除也有外键级联保护。

导出的 Session 文件包含登录凭据，应按敏感文件保管，不要提交到 Git 或公开附件。系统优先比对 Session `user.id` 与账号 `chatgpt_user_id`，缺失 ID 时比对邮箱；无法确认身份或身份不一致时拒绝导出。此校验确认用户身份，不保证 Session 属于某个特定工作区。

服务端使用账号配置的代理访问固定的 ChatGPT 地址，不跟随重定向；代理查询失败时不改走直连。单次上游请求超时 20 秒，响应上限 1 MiB。Cookie 过期、未登录、Cloudflare/浏览器验证、代理不可用或无效响应都会报错。保存 Cookie 不会消除上游风控要求，也不会自动续期网页登录。关闭窗口、切换账号或离开页面后，晚到响应不会下载文件或重新弹出二次验证。

## 管理 API

- `POST /api/v1/admin/accounts/:id/export-session`：请求 `{}` 使用已保存 Cookie；`{"cookie":"<完整 Cookie>","remember_cookie":true}` 验证后加密保存并导出。`remember_cookie` 默认 `false`，兼容原来的单次 Cookie 导出。标准 API envelope 的 `data` 为真实上游 Session JSON，不修改账号 OAuth token。
- `GET /api/v1/admin/accounts/:id/session-cookie`：返回 `data: {"configured":true,"storage_enabled":true}` 等状态。
- `DELETE /api/v1/admin/accounts/:id/session-cookie`：删除已保存 Cookie，返回 `data: {"deleted":true}`；无需解密，因此密钥丢失时仍可删除。

这些接口要求现有管理面认证，观察员仅可操作授权分组内的账号。导出/保存与删除要求现有 step-up 策略；状态查询只返回布尔值。响应均设置 `Cache-Control: no-store`。

新增迁移 `285_openai_session_cookies.sql`，不修改现有账号 JSON 格式。回滚应用代码后可保留新增表，旧版本忽略该表；不会自动删除已保存密文。

## 界面截图

以下是使用实际组件、虚构账号与模拟 API 响应的本地离线预览，未进行真实 ChatGPT 登录联调。

修改前：

[修改前的 OpenAI 账号菜单](https://github.com/ranxi2001/sub2api/blob/297e9201a2f6d6fe70073b370a5b71f49579accf/docs/images/openai-session-export-before.png)

新增导出与设置入口：

[新增 Session 导出与设置入口](https://github.com/ambi-vip/sub2api/blob/c0f49b8b3fe190d75ce506e33ff0448d9805bc9c/docs/images/openai-session-export-menu.png)

首次配置：

[Session 首次配置](https://github.com/ambi-vip/sub2api/blob/c0f49b8b3fe190d75ce506e33ff0448d9805bc9c/docs/images/openai-session-export-dialog.png)

已保存凭据的管理界面：

[Session 登录设置](https://github.com/ambi-vip/sub2api/blob/c0f49b8b3fe190d75ce506e33ff0448d9805bc9c/docs/images/openai-session-export-saved.png)
