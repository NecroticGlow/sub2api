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

func deepSeekCacheFallbackTestContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(deepSeekCacheEstimateContextKey, deepSeekCacheCandidate{
		commonBytes: 1000, currentBytes: 1000, confidencePercent: 65,
	})
	return c
}

func TestScanCCStreamAppliesDeepSeekCacheEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := deepSeekCacheFallbackTestContext()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(
		"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":1000,\"completion_tokens\":5}}\n\n" +
			"data: [DONE]\n\n",
	))}

	state := (&OpenAIGatewayService{}).scanCCStream(
		c, resp, "test", "request-id", time.Now(), func(_ *apicompat.ChatCompletionsChunk) {},
	)
	require.Equal(t, 566, state.Usage.CacheReadInputTokens)
	require.Equal(t, 1000, state.Usage.InputTokens)
	require.True(t, state.SawDone)
}

func TestReadCCUpstreamJSONResponseAppliesDeepSeekCacheEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := deepSeekCacheFallbackTestContext()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(
		`{"choices":[],"usage":{"prompt_tokens":1000,"completion_tokens":5}}`,
	))}
	writeError := func(_ *gin.Context, _ int, _, _ string) {}

	_, usage, err := (&OpenAIGatewayService{}).readCCUpstreamJSONResponse(c, resp, writeError)
	require.NoError(t, err)
	require.Equal(t, 566, usage.CacheReadInputTokens)
	require.Equal(t, 1000, usage.InputTokens)
}
