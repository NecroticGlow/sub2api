-- 把 deepseek 平台加入 user_platform_quotas.platform 与
-- composite_model_routes.target_platform 的 CHECK 约束。
--
-- 背景：新增 DeepSeek 官方 API 渠道（internal/domain/constants.go 的
-- PlatformDeepSeek）。与 157（grok）同理：注册时 snapshotPlatformQuotaDefaults
-- 会写入 deepseek 默认配额行，composite 路由也可指向 deepseek，
-- 两个 CHECK 必须与代码平台列表对齐。
--
-- DROP ... IF EXISTS 保证可重入；新约束是旧约束的超集，存量行瞬时校验通过。
ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'deepseek'));

ALTER TABLE composite_model_routes
    DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check;

ALTER TABLE composite_model_routes
    ADD CONSTRAINT composite_model_routes_target_platform_check
    CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'deepseek'));
