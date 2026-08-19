//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeepSeekMonitorExtractsReasoningContentWhenContentIsEmpty(t *testing.T) {
	resp := []byte(`{"choices":[{"message":{"content":"","reasoning_content":"The answer is 29"}}]}`)
	require.Equal(t, "The answer is 29", extractOpenAIChatText(resp))
}

func TestDeepSeekMonitorUsesOpenAICompatibleAdapter(t *testing.T) {
	adapter, mode, ok := providerAdapterFor(MonitorProviderDeepSeek, MonitorAPIModeResponses)
	require.True(t, ok)
	require.Equal(t, MonitorAPIModeResponses, mode)
	body, err := adapter.buildBody("deepseek-v4-flash", "Reply with exactly: 29")
	require.NoError(t, err)
	require.Contains(t, string(body), "deepseek-v4-flash")
}

func TestDeepSeekMonitorUpdatePreservesResponsesMode(t *testing.T) {
	deepseek := MonitorProviderDeepSeek
	existing := &ChannelMonitor{
		Provider:        MonitorProviderOpenAI,
		APIMode:         MonitorAPIModeResponses,
		PrimaryModel:    "deepseek-v4-flash",
		IntervalSeconds: 60,
	}

	require.NoError(t, applyMonitorUpdate(existing, ChannelMonitorUpdateParams{Provider: &deepseek}))
	require.Equal(t, MonitorProviderDeepSeek, existing.Provider)
	require.Equal(t, MonitorAPIModeResponses, existing.APIMode)
}
