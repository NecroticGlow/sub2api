//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

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
		{"reference with prompt boilerplate", "Do not browse the web or speculate. If you are uncertain, say uncertain.\n" + referenceIntelligenceAnswer, "passed"},
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

func TestV2811IntelligenceQuestions(t *testing.T) {
	status, checks := assessIntelligenceResponseForQuestion(IntelligenceQuestionCandy, "答案是 21")
	require.Equal(t, "passed", status)
	require.Equal(t, "candy", checks[0].Item)
	status, _ = assessIntelligenceResponseForQuestion(IntelligenceQuestionCandy, "答案是 20")
	require.Equal(t, "manual_review", status)
	status, checks = assessIntelligenceResponseForQuestion(IntelligenceQuestionPelican, "<!doctype html><svg></svg>")
	require.Equal(t, "passed", status)
	require.Equal(t, "pelican", checks[0].Item)
	status, _ = assessIntelligenceResponseForQuestion(IntelligenceQuestionPelican, "I cannot create HTML")
	require.Equal(t, "manual_review", status)
}

func TestIntelligenceHistoryPersistsServerSide(t *testing.T) {
	svc, _, account := intelligenceTestFixture()
	repo := svc.accountRepo.(*openAIAccountTestRepo)
	result := &IntelligenceTestResult{AccountID: account.ID, Model: IntelligenceTestModel, Status: "passed", ResponseText: referenceIntelligenceAnswer, TestedAt: time.Now().UTC()}
	require.NoError(t, svc.SaveIntelligenceTestResult(context.Background(), result))
	raw, ok := repo.updatedExtra[intelligenceTestHistoryExtraKey]
	require.True(t, ok)
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	var history []IntelligenceTestResult
	require.NoError(t, json.Unmarshal(encoded, &history))
	require.Len(t, history, 1)
	require.Equal(t, "passed", history[0].Status)
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
	upstream.responses = []*http.Response{newJSONResponse(200, "data: "+string(delta)+"\n\ndata: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":100,\"output_tokens\":20,\"total_tokens\":120,\"input_tokens_details\":{\"cached_tokens\":25}}}}\n\n")}
	result, err := svc.TestAccountIntelligence(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, "passed", result.Status)
	require.Equal(t, referenceIntelligenceAnswer, result.ResponseText)
	require.Equal(t, &TestUsage{InputTokens: 100, OutputTokens: 20, TotalTokens: 120, CachedInputTokens: 25}, result.Usage)
	require.InDelta(t, 0.001775, result.Cost.TotalCostUSD, 1e-12)
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

func TestIntelligenceCandyQuestionUsesV2811PromptAndReasoning(t *testing.T) {
	svc, upstream, account := intelligenceTestFixture()
	delta, err := json.Marshal(map[string]string{"type": "response.output_text.delta", "delta": "答案是 21"})
	require.NoError(t, err)
	upstream.responses = []*http.Response{newJSONResponse(200, "data: "+string(delta)+"\n\ndata: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":2,\"total_tokens\":12}}}\n\n")}
	result, err := svc.TestAccountIntelligence(context.Background(), account.ID, IntelligenceQuestionCandy, "high")
	require.NoError(t, err)
	require.Equal(t, "passed", result.Status)
	require.Equal(t, IntelligenceQuestionCandy, result.QuestionKind)
	body, err := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	input := payload["input"].([]any)
	text := input[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"]
	require.Equal(t, IntelligenceCandyPrompt, text)
	require.Equal(t, map[string]any{"effort": "high"}, payload["reasoning"])
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
