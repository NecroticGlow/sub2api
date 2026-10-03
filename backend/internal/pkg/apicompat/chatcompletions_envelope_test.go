package apicompat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnwrapChatCompletionsResponse(t *testing.T) {
	inner := `{"model":"deepseek-v4.1-flash","choices":[{"message":{"content":"9","reasoning":"private","tool_calls":[]},"finish_reason":"stop"}],"usage":{"prompt_tokens":80,"completion_tokens":23,"prompt_tokens_details":{"cached_tokens":40}},"provider_extension":{"retained":true}}`
	require.Equal(t, inner, string(UnwrapChatCompletionsResponse([]byte(`{"data":`+inner+`,"success":true}`))))
	for _, body := range []string{
		inner,
		`{"choices":[],"data":` + inner + `,"success":true}`,
		`{"choices":null,"data":` + inner + `,"success":true}`,
		`{"error":{"message":"failed"},"data":` + inner + `,"success":true}`,
		`{"data":` + inner + `,"success":false}`,
		`{"data":` + inner + `}`,
		`{"success":true,"data":{"choices":{},"content":"not a completion"}}`,
		`{"success":true,"data":{"error":{},"choices":[]}}`,
		`{"success":true,"data":{"images":[]}}`,
		`{"success":true,"data":null}`,
		`{"success":true,"data":`,
		`data: {"choices":[]}\n\n`,
	} {
		require.Equal(t, body, string(UnwrapChatCompletionsResponse([]byte(body))), body)
	}
}
