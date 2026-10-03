package apicompat

import (
	"bytes"
	"encoding/json"
)

// UnwrapChatCompletionsResponse accepts the successful data envelope used by
// some compatible providers (including ClinePass). Preserve the inner JSON
// verbatim, including usage, reasoning, tool calls and provider extensions.
// Standard responses and errors must never be replaced by nested data.
func UnwrapChatCompletionsResponse(body []byte) []byte {
	var envelope map[string]json.RawMessage
	if json.Unmarshal(body, &envelope) != nil {
		return body
	}
	if _, exists := envelope["choices"]; exists {
		return body
	}
	if _, exists := envelope["error"]; exists || !bytes.Equal(bytes.TrimSpace(envelope["success"]), []byte("true")) {
		return body
	}
	var data map[string]json.RawMessage
	if json.Unmarshal(envelope["data"], &data) != nil {
		return body
	}
	if _, exists := data["error"]; exists {
		return body
	}
	choices := bytes.TrimSpace(data["choices"])
	if len(choices) == 0 || choices[0] != '[' {
		return body
	}
	return envelope["data"]
}
