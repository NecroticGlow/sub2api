//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const clineEnvelopeFixture = `{"success":true,"data":{"id":"chatcmpl_wrapped","object":"chat.completion","model":"cline-pass/deepseek-v4.1-flash","choices":[{"index":0,"message":{"role":"assistant","content":"9","reasoning":"private thinking","provider_metadata":{"retained":true}},"finish_reason":"stop"}],"usage":{"prompt_tokens":80,"completion_tokens":23,"completion_tokens_details":{"reasoning_tokens":20},"prompt_tokens_details":{"cached_tokens":40}},"provider_extension":{"retained":true}}}`

func TestChatEnvelopeRawForward(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"deepseek-v4.1-flash","messages":[{"role":"user","content":"4+5?"}],"stream":false}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(clineEnvelopeFixture))}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	account := rawChatCompletionsTestAccount()
	account.Platform = PlatformDeepseek
	account.Credentials["model_mapping"] = map[string]any{"deepseek-v4.1-flash": "cline-pass/deepseek-v4.1-flash"}
	result, err := svc.forwardAsRawChatCompletions(context.Background(), c, account, body, "")
	require.NoError(t, err)
	require.Equal(t, "9", gjson.Get(rec.Body.String(), "choices.0.message.content").String())
	require.Equal(t, "private thinking", gjson.Get(rec.Body.String(), "choices.0.message.reasoning").String())
	require.Equal(t, "deepseek-v4.1-flash", gjson.Get(rec.Body.String(), "model").String())
	require.True(t, gjson.Get(rec.Body.String(), "provider_extension.retained").Bool())
	require.False(t, gjson.Get(rec.Body.String(), "data").Exists())
	require.Equal(t, 80, result.Usage.InputTokens)
	require.Equal(t, 23, result.Usage.OutputTokens)
	require.Equal(t, 40, result.Usage.CacheReadInputTokens)
}

func TestChatEnvelopeProtocolConversions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"responses", "messages"} {
		t.Run(protocol, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, nil)
			resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(clineEnvelopeFixture))}
			svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig()}
			var result *OpenAIForwardResult
			var err error
			if protocol == "responses" {
				result, err = svc.bufferChatCompletionsAsResponses(c, resp, "deepseek-v4.1-flash", nil, nil, false, nil, "deepseek-v4.1-flash", "cline-pass/deepseek-v4.1-flash", nil, nil, false, time.Now(), false)
			} else {
				result, err = svc.bufferChatCompletionsAsAnthropic(c, resp, "deepseek-v4.1-flash", "deepseek-v4.1-flash", "cline-pass/deepseek-v4.1-flash", nil, nil, time.Now())
			}
			require.NoError(t, err)
			require.Equal(t, 80, result.Usage.InputTokens)
			require.Equal(t, 23, result.Usage.OutputTokens)
			require.Equal(t, 40, result.Usage.CacheReadInputTokens)
			if protocol == "responses" {
				require.Contains(t, gjson.Get(rec.Body.String(), "output.#.content.#.text").String(), "9")
			} else {
				require.Contains(t, gjson.Get(rec.Body.String(), "content.#.text").String(), "9")
			}
		})
	}
}

func TestChatEnvelopeMonitor(t *testing.T) {
	text := extractMonitorResponseText(providerDeepseekChatAdapter, []byte(clineEnvelopeFixture))
	require.Equal(t, "9", text)
	require.True(t, validateChallenge(text, "9"))
	for _, body := range []string{
		`{"success":false,"data":{"choices":[{"message":{"content":"9"}}]}}`,
		`{"success":true,"data":{"choices":[{"message":{"content":"","reasoning":"9"}}]}}`,
	} {
		require.False(t, validateChallenge(extractMonitorResponseText(providerDeepseekChatAdapter, []byte(body)), "9"))
	}
}

func TestChatEnvelopeAccountTest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	err := (&AccountTestService{}).processOpenAIChatCompletionsStream(c, strings.NewReader("data: "+clineEnvelopeFixture+"\n\ndata: [DONE]\n\n"))
	require.NoError(t, err)
	require.Contains(t, rec.Body.String(), `"type":"content","text":"9"`)
	require.Contains(t, rec.Body.String(), `"success":true`)
}
