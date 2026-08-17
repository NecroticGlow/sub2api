// Package deepseek holds constants for the DeepSeek official API platform.
//
// DeepSeek exposes an OpenAI-compatible Chat Completions API. It has no
// /v1/responses endpoint, so gateway traffic for DeepSeek accounts always
// uses the raw Chat Completions forwarding path.
package deepseek

// DefaultBaseURL is the official DeepSeek API endpoint. The gateway appends
// /v1/chat/completions (or /chat/completions when the stored base URL already
// ends with a version segment).
const DefaultBaseURL = "https://api.deepseek.com"

// ClinePassBaseURL is the Cline API endpoint (OpenAI-compatible Chat
// Completions). ClinePass serves DeepSeek models under the
// "cline-pass/deepseek-v4-*" slugs; billing still resolves them to the
// official DeepSeek prices (see BillingService.getFallbackPricing).
const ClinePassBaseURL = "https://api.cline.bot/api/v1"

// Model IDs of the official DeepSeek text models.
// deepseek-chat / deepseek-reasoner are the official compatibility aliases
// that map onto the V4 series upstream.
const (
	DefaultChatModelID     = "deepseek-chat"
	DefaultReasonerModelID = "deepseek-reasoner"
	V4ProModelID           = "deepseek-v4-pro"
	V4FlashModelID         = "deepseek-v4-flash"
)

// ClinePass model slugs for the DeepSeek models it serves.
const (
	ClinePassV4ProModelID   = "cline-pass/deepseek-v4-pro"
	ClinePassV4FlashModelID = "cline-pass/deepseek-v4-flash"
)

// ClinePassModelIDs lists every model slug served by ClinePass
// (https://docs.cline.bot/getting-started/clinepass). DeepSeek-platform
// accounts pointed at the ClinePass upstream can map/whitelist any of these;
// billing resolves each slug to the underlying vendor's official rates.
func ClinePassModelIDs() []string {
	return []string{
		ClinePassV4ProModelID,
		ClinePassV4FlashModelID,
		"cline-pass/glm-5.2",
		"cline-pass/kimi-k3",
		"cline-pass/kimi-k2.7-code",
		"cline-pass/kimi-k2.6",
		"cline-pass/mimo-v2.5",
		"cline-pass/mimo-v2.5-pro",
		"cline-pass/minimax-m3",
		"cline-pass/qwen3.8-max",
		"cline-pass/qwen3.7-max",
		"cline-pass/qwen3.7-plus",
	}
}

// Model mirrors the OpenAI models-list entry shape.
type Model struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

// DefaultModels is the DeepSeek models list served when no account-level
// model mapping narrows it down.
var DefaultModels = []Model{
	{ID: DefaultChatModelID, Object: "model", Created: 1735689600, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek Chat"},
	{ID: DefaultReasonerModelID, Object: "model", Created: 1735689600, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek Reasoner"},
	{ID: V4ProModelID, Object: "model", Created: 1767225600, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek V4 Pro"},
	{ID: V4FlashModelID, Object: "model", Created: 1767225600, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek V4 Flash"},
}

// DefaultModelIDs returns the default model ID list.
func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, model := range DefaultModels {
		ids[i] = model.ID
	}
	return ids
}
