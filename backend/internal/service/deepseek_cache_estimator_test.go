package service

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOllamaCloudAccountRequiresExactHost(t *testing.T) {
	require.True(t, ollamaCloudAccount(&Account{Credentials: map[string]any{"base_url": "https://ollama.com/v1"}}))
	require.True(t, ollamaCloudAccount(&Account{Credentials: map[string]any{"base_url": "HTTPS://WWW.OLLAMA.COM:443/v1"}}))
	require.False(t, ollamaCloudAccount(&Account{Credentials: map[string]any{"base_url": "https://ollama.com.evil.test/v1"}}))
}

func TestDeepSeekCacheGroup(t *testing.T) {
	groups := [][]int64{{971, 979}, {990}}
	require.Equal(t, deepSeekCacheGroup(971, groups), deepSeekCacheGroup(979, groups))
	require.NotEqual(t, deepSeekCacheGroup(971, groups), deepSeekCacheGroup(990, groups))
}

func TestDeepSeekCacheEstimatorPrepareAndApply(t *testing.T) {
	gin.SetMode(gin.TestMode)
	estimator := newDeepSeekCacheEstimator(nil)
	estimator.config = deepSeekCacheEstimateSettings{Enabled: true, AccountGroups: [][]int64{{971, 979}, {990}}, TTLSeconds: 900, BlockBytes: 64, MaxFingerprints: 16}
	estimator.loadedAt = time.Now()
	prefix := "This is a deliberately long stable system prompt used to exercise block matching."
	first := []byte(`{"model":"deepseek-v4-flash:0731","instructions":"` + prefix + `","input":[{"role":"user","content":"one"}]}`)
	second := []byte(`{"model":"deepseek-v4-flash:0731","instructions":"` + prefix + `","input":[{"role":"user","content":"one"},{"role":"user","content":"two"}]}`)
	c1, _ := gin.CreateTestContext(nil)
	estimator.prepare(context.Background(), c1, &Account{ID: 971, Credentials: map[string]any{"base_url": "https://ollama.com"}}, "deepseek-v4-flash:0731", first)
	c2, _ := gin.CreateTestContext(nil)
	estimator.prepare(context.Background(), c2, &Account{ID: 979, Credentials: map[string]any{"base_url": "https://ollama.com/v1"}}, "deepseek-v4-flash:0731", second)
	out, estimated := applyDeepSeekCacheEstimate(c2, []byte(`{"usage":{"input_tokens":1200,"output_tokens":5,"input_tokens_details":{"cached_tokens":0}}}`))
	require.Positive(t, estimated)
	require.Equal(t, int64(estimated), gjson.GetBytes(out, "usage.input_tokens_details.cached_tokens").Int())
}

func TestDeepSeekCacheEstimatorKeepsLargePromptAcrossUnrelatedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	estimator := newDeepSeekCacheEstimator(nil)
	estimator.config = deepSeekCacheEstimateSettings{
		Enabled: true, AccountGroups: [][]int64{{971, 979}}, TTLSeconds: 900,
		BlockBytes: 64, MaxFingerprints: 4,
	}
	estimator.loadedAt = time.Now()
	account := &Account{ID: 979, Credentials: map[string]any{"base_url": "https://ollama.com"}}
	stable := "This stable prefix is intentionally longer than one fingerprint block so it can be matched."
	largeFirst := []byte(`{"instructions":"` + stable + `","messages":[{"role":"user","content":"first"}]}`)
	unrelated := []byte(`{"messages":[{"role":"user","content":"unrelated short request"}]}`)
	largeNext := []byte(`{"instructions":"` + stable + `","messages":[{"role":"user","content":"first"},{"role":"user","content":"next"}]}`)

	seed, _ := gin.CreateTestContext(nil)
	estimator.prepare(context.Background(), seed, account, "deepseek-v4-flash:0731", largeFirst)
	noise, _ := gin.CreateTestContext(nil)
	estimator.prepare(context.Background(), noise, account, "deepseek-v4-flash:0731", unrelated)
	current, _ := gin.CreateTestContext(nil)
	estimator.prepare(context.Background(), current, account, "deepseek-v4-flash:0731", largeNext)

	out, estimated := applyDeepSeekCacheEstimate(current, []byte(`{"usage":{"prompt_tokens":800000,"completion_tokens":5}}`))
	require.Positive(t, estimated)
	require.Greater(t, gjson.GetBytes(out, "usage.prompt_tokens_details.cached_tokens").Int(), int64(100000))
}

func TestApplyDeepSeekCacheEstimateChatCompletionsUsage(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	c.Set(deepSeekCacheEstimateContextKey, deepSeekCacheCandidate{
		commonBytes: 900, currentBytes: 1000, confidencePercent: 85,
	})
	out, estimated := applyDeepSeekCacheEstimate(c, []byte(`{"usage":{"prompt_tokens":2000,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":0}}}`))
	require.Equal(t, 1421, estimated)
	require.Equal(t, int64(estimated), gjson.GetBytes(out, "usage.prompt_tokens_details.cached_tokens").Int())
}
