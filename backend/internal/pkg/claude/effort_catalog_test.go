package claude

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEffortLevelsForModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		model string
		want  []string
	}{
		{model: "claude-opus-4-6", want: []string{"low", "medium", "high", "max"}},
		{model: "anthropic/claude-sonnet-4-6", want: []string{"low", "medium", "high", "max"}},
		{model: "claude-opus-5", want: []string{"low", "medium", "high", "xhigh", "max"}},
		{model: "anthropic/claude-opus-5.5", want: []string{"low", "medium", "high", "xhigh", "max"}},
		{model: "claude-opus-4-5-20251101", want: []string{"low", "medium", "high"}},
		{model: "claude-haiku-4-5-20251001", want: nil},
		{model: "gpt-5.6", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, EffortLevelsForModel(tt.model))
		})
	}
}

func TestSupportsAdaptiveThinking(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"claude-opus-5-5", "claude-opus-4-6", "claude-sonnet-4-6", "claude-opus-4-8-20260501", "anthropic/claude-opus-5.5"} {
		require.True(t, SupportsAdaptiveThinking(model), model)
	}
	// Opus 4.5 takes effort but needs a thinking budget; the rest take neither.
	for _, model := range []string{"claude-opus-4-5", "claude-opus-4-5-20251101", "claude-haiku-4-5-20251001", "claude-sonnet-4-5", "gpt-6-astra"} {
		require.False(t, SupportsAdaptiveThinking(model), model)
	}
}

func TestIsOpus55OpenRouterExactAlias(t *testing.T) {
	t.Parallel()
	for _, model := range []string{"claude-opus-5-5", "anthropic/claude-opus-5.5"} {
		require.True(t, IsOpus55(model), model)
	}
	for _, model := range []string{"claude-opus-5", "anthropic/claude-opus-5.6", "anthropic/claude-opus-5.5-preview"} {
		require.False(t, IsOpus55(model), model)
	}
}
