-- Cache billing now uses upstream-reported usage only. Historical usage logs
-- remain intact; remove only the obsolete estimation setting.
DELETE FROM settings WHERE key = 'deepseek_cache_estimate';
