//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const referenceIntelligenceAnswer = `| iPhone | **iPhone 17 generation**, including iPhone 17 Pro and Pro Max |
| Apple's official announcement date | **September 9, 2025** |
| Official on-sale date | **September 19, 2025** |
| NVIDIA GPU | **GeForce RTX 5090** (consumer graphics) |
| Android | **Android 16** |
| macOS | **macOS Tahoe 26** |
| Windows minor/feature version | **Windows 11, version 25H2** |`

func TestIntelligenceAssessment(t *testing.T) {
	for _, tc := range []struct{ name, answer, status string }{
		{"reference", referenceIntelligenceAnswer, "passed"},
		{"case and ISO dates", strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(referenceIntelligenceAnswer), "SEPTEMBER 9, 2025", "2025-09-09"), "SEPTEMBER 19, 2025", "2025-09-19"), "passed"},
		{"older Android", strings.ReplaceAll(referenceIntelligenceAnswer, "Android 16", "Android 15"), "manual_review"},
		{"uncertain", referenceIntelligenceAnswer + "\nI am uncertain about this.", "manual_review"},
		{"negated", referenceIntelligenceAnswer + "\nThese are not my answers.", "manual_review"},
		{"empty", "", "manual_review"},
		{"conflicting answer", referenceIntelligenceAnswer + "\nThe latest is iPhone 18.", "manual_review"},
		{"swapped dates", strings.NewReplacer("September 9, 2025", "September 19, 2025", "September 19, 2025", "September 9, 2025").Replace(referenceIntelligenceAnswer), "manual_review"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, checks := assessIntelligenceResponse(tc.answer)
			require.Equal(t, tc.status, status)
			require.Len(t, checks, 7)
		})
	}
}

func intelligenceTestFixture() (*AccountTestService, *queuedHTTPUpstream, *Account) {
	account := &Account{ID: 99123, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1, Credentials: map[string]any{"access_token": "test-token", "model_mapping": map[string]any{IntelligenceTestModel: "gpt-5.4"}}}
	repo := &openAIAccountTestRepo{mockAccountRepoForGemini: mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}}
	upstream := &queuedHTTPUpstream{}
	return &AccountTestService{accountRepo: repo, httpUpstream: upstream}, upstream, account
}

func TestIntelligenceUsesExactPromptModelAndNoOverdraft(t *testing.T) {
	svc, upstream, account := intelligenceTestFixture()
	coordinator := &accountTestOverdraftCoordinatorStub{}
	svc.cfg = &config.Config{}
	svc.cfg.Gateway.CodexQuotaOverdraftEnabled = true
	svc.codexQuotaOverdraft = coordinator
	delta, err := json.Marshal(map[string]string{"type": "response.output_text.delta", "delta": referenceIntelligenceAnswer})
	require.NoError(t, err)
	upstream.responses = []*http.Response{newJSONResponse(200, "data: "+string(delta)+"\n\ndata: {\"type\":\"response.completed\"}\n\n")}
	result, err := svc.TestAccountIntelligence(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, "passed", result.Status)
	require.Equal(t, referenceIntelligenceAnswer, result.ResponseText)
	require.Len(t, upstream.requests, 1)
	request := upstream.requests[0]
	require.Equal(t, chatgptCodexAPIURL, request.URL.String())
	body, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, IntelligenceTestModel, payload["model"])
	require.Equal(t, "none", payload["tool_choice"])
	require.Empty(t, payload["tools"])
	input := payload["input"].([]any)
	require.Len(t, input, 1)
	text := input[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"]
	require.Equal(t, IntelligenceTestPrompt, text)
	require.NotContains(t, string(body), "September 9")
	require.Zero(t, coordinator.observeCalls)
	require.Zero(t, coordinator.businessCalls)
}

func TestIntelligenceRejectsOtherAccountsAndDuplicateTests(t *testing.T) {
	svc, upstream, account := intelligenceTestFixture()
	account.Type = AccountTypeAPIKey
	_, err := svc.TestAccountIntelligence(context.Background(), account.ID)
	require.ErrorIs(t, err, ErrIntelligenceAccountType)
	account.Type, account.Platform = AccountTypeOAuth, PlatformAnthropic
	_, err = svc.TestAccountIntelligence(context.Background(), account.ID)
	require.ErrorIs(t, err, ErrIntelligenceAccountType)
	account.Platform = PlatformOpenAI
	intelligenceTestsRunning.Store(account.ID, true)
	defer intelligenceTestsRunning.Delete(account.ID)
	_, err = svc.TestAccountIntelligence(context.Background(), account.ID)
	require.ErrorIs(t, err, ErrIntelligenceTestBusy)
	require.Empty(t, upstream.requests)
}

func TestIntelligenceFailedOrTruncatedStreamNeverPasses(t *testing.T) {
	for _, body := range []string{"data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n", "data: {\"type\":\"response.failed\"}\n\n", "data: {\"type\":\"response.completed\"}\n\n"} {
		svc, upstream, account := intelligenceTestFixture()
		upstream.responses = []*http.Response{newJSONResponse(200, body)}
		result, err := svc.TestAccountIntelligence(context.Background(), account.ID)
		require.NoError(t, err)
		require.Equal(t, "error", result.Status)
		require.NotEmpty(t, result.Error)
	}
}
