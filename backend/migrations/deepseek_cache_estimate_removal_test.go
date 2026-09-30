package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDeepSeekCacheEstimateSettingOnly(t *testing.T) {
	content, err := FS.ReadFile("263_remove_deepseek_cache_estimate.sql")
	require.NoError(t, err)
	sql := string(content)
	require.Contains(t, sql, "DELETE FROM settings WHERE key = 'deepseek_cache_estimate';")
	require.NotContains(t, sql, "UPDATE usage_logs")
	require.NotContains(t, sql, "DELETE FROM usage_logs")
	require.NotContains(t, sql, "DROP TABLE")
}
