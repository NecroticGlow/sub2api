package ent_test

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/group"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/ent/schema"
	"github.com/stretchr/testify/require"
)

// Blank-importing runtime catches generated descriptor index drift at startup,
// before an auto-merged schema can be shipped with mismatched generated defaults.
func TestMergedGroupDefaultsMatchSchema(t *testing.T) {
	defaults := map[string]any{
		"rpm_limit":                       group.DefaultRpmLimit,
		"user_concurrency_limit":          group.DefaultUserConcurrencyLimit,
		"max_reasoning_effort":            group.DefaultMaxReasoningEffort,
		"max_reasoning_effort_over_limit": group.DefaultMaxReasoningEffortOverLimit,
	}
	for _, field := range (schema.Group{}).Fields() {
		desc := field.Descriptor()
		if actual, ok := defaults[desc.Name]; ok {
			require.Equal(t, desc.Default, actual, desc.Name)
			delete(defaults, desc.Name)
		}
	}
	require.Empty(t, defaults)
	require.Equal(t, 5, account.DefaultRateLimit429RetryCount)
}
