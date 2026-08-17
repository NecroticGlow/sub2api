-- Add DeepSeek to passive channel-monitor V2's factory platform scope.
-- Existing operator customizations are preserved: only append the platform
-- when it is not already present, and never replace the whole JSON document.

UPDATE channel_monitor_v2_config
SET
    platforms = platforms || '[{"platform":"deepseek","enabled":true,"models":["deepseek-v4-flash","deepseek-v4-pro"]}]'::jsonb,
    version = version + 1,
    updated_at = NOW()
WHERE id = 1
  AND NOT EXISTS (
    SELECT 1
    FROM jsonb_array_elements(platforms) AS p
    WHERE p->>'platform' = 'deepseek'
  );
