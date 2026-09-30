package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCCPipelineUsesOnlyReportedCacheUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name  string
		usage string
		cache int
	}{
		{"missing cache stays zero", `{"prompt_tokens":1000,"completion_tokens":5}`, 0},
		{"explicit miss stays zero", `{"prompt_tokens":1000,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":0}}`, 0},
		{"reported hits preserved", `{"prompt_tokens":1000,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":321}}`, 321},
		{"native DeepSeek hits preserved", `{"prompt_tokens":1000,"completion_tokens":5,"prompt_cache_hit_tokens":321,"prompt_cache_miss_tokens":679}`, 321},
		{"native DeepSeek miss stays zero", `{"prompt_tokens":1000,"completion_tokens":5,"prompt_cache_hit_tokens":0,"prompt_cache_miss_tokens":1000}`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"choices":[],"usage":` + tc.usage + `}`
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			resp := &http.Response{Body: io.NopCloser(strings.NewReader("data: " + body + "\n\ndata: [DONE]\n\n"))}
			state := (&OpenAIGatewayService{}).scanCCStream(c, resp, "test", "request-id", time.Now(), func(_ *apicompat.ChatCompletionsChunk) {})
			require.Equal(t, tc.cache, state.Usage.CacheReadInputTokens)
			require.Equal(t, 1000, state.Usage.InputTokens)
			require.True(t, state.SawDone)

			resp = &http.Response{Body: io.NopCloser(strings.NewReader(body))}
			_, usage, err := (&OpenAIGatewayService{}).readCCUpstreamJSONResponse(c, resp, func(_ *gin.Context, _ int, _, _ string) {})
			require.NoError(t, err)
			require.Equal(t, tc.cache, usage.CacheReadInputTokens)
			require.Equal(t, 1000, usage.InputTokens)
		})
	}
}
