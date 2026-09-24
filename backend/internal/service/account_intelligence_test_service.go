package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const IntelligenceTestModel = "gpt-6-astra"
const IntelligenceTestPrompt = `Do not browse the web or speculate. Based only on your existing knowledge, tell me the latest-generation iPhone model you explicitly know about, along with Apple’s official announcement date and the official on-sale date.

Also provide the latest NVIDIA GPU model, Android version, macOS version, and Windows minor version you explicitly know about.

If you are uncertain about any item, simply say “uncertain” for that item. Do not guess or fill in missing information.`

var ErrIntelligenceAccountType = errors.New("intelligence tests require an OpenAI OAuth account")
var ErrIntelligenceTestBusy = errors.New("a test is already running for this account")
var intelligenceTestsRunning sync.Map

type intelligenceTestContextKey struct{}

type IntelligenceCheck struct {
	Item     string `json:"item"`
	Expected string `json:"expected"`
	Matched  bool   `json:"matched"`
}

type IntelligenceTestResult struct {
	AccountID          int64                 `json:"account_id"`
	Model              string                `json:"model"`
	Status             string                `json:"status"`
	ResponseText       string                `json:"response_text"`
	Error              string                `json:"error,omitempty"`
	Checks             []IntelligenceCheck   `json:"checks"`
	TestedAt           time.Time             `json:"tested_at"`
	LatencyMS          int64                 `json:"latency_ms"`
	Usage              *TestUsage            `json:"usage,omitempty"`
	Cost               *IntelligenceTestCost `json:"cost,omitempty"`
	CurrentConcurrency int                   `json:"current_concurrency"`
}

type IntelligenceTestCost struct {
	InputCostUSD       float64 `json:"input_cost_usd"`
	CachedInputCostUSD float64 `json:"cached_input_cost_usd"`
	OutputCostUSD      float64 `json:"output_cost_usd"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
	PricingNote        string  `json:"pricing_note"`
}

// Deliberately conservative: lexical agreement is a screening signal, not proof
// of model identity or capability. Every ambiguous result requires human review.
var intelligenceReference = []struct {
	item, expected string
	pattern        *regexp.Regexp
}{
	{"iphone", "iPhone 17 generation, including iPhone 17 Pro and Pro Max", regexp.MustCompile(`(?i)\biphone\s+17\b`)},
	{"announcement", "September 9, 2025", regexp.MustCompile(`(?i)\b(?:september\s+9(?:th)?[,]?\s+2025|9\s+september[,]?\s+2025|2025-09-09)\b`)},
	{"on_sale", "September 19, 2025", regexp.MustCompile(`(?i)\b(?:september\s+19(?:th)?[,]?\s+2025|19\s+september[,]?\s+2025|2025-09-19)\b`)},
	{"nvidia", "GeForce RTX 5090", regexp.MustCompile(`(?i)\b(?:geforce\s+)?rtx\s+5090\b`)},
	{"android", "Android 16", regexp.MustCompile(`(?i)\bandroid\s+16\b`)},
	{"macos", "macOS Tahoe 26", regexp.MustCompile(`(?i)\bmacos\s+tahoe\s+26\b`)},
	{"windows", "Windows 11, version 25H2", regexp.MustCompile(`(?i)\bwindows\s+11\b[^\n.!?]{0,60}\b25h2\b`)},
}
var intelligenceAmbiguous = regexp.MustCompile(`(?i)\b(uncertain|unsure|unknown|cannot|can't|might|maybe|possibly|guess|speculat\w*)\b|不确定|不清楚|无法确认`)
var intelligenceNegated = regexp.MustCompile(`(?i)\b(?:not\s+(?:my|the|an?)\s+answers?|answer(?:s)?\s+(?:are|is)\s+not\s+correct|do\s+not\s+match)\b|不是答案|回答不正确`)
var intelligencePro = regexp.MustCompile(`(?i)\b(?:iphone\s+17\s+)?pro\b`)
var intelligenceProMax = regexp.MustCompile(`(?i)\bpro\s+max\b`)
var intelligenceAnnouncement = regexp.MustCompile(`(?i)announc\w*|发布`)
var intelligenceOnSale = regexp.MustCompile(`(?i)on[ -]sale|availability|available|发售|上市`)
var intelligenceVersionClaims = []struct {
	pattern  *regexp.Regexp
	expected string
}{
	{regexp.MustCompile(`(?i)\biphone\s+(\d+)\b`), "17"},
	{regexp.MustCompile(`(?i)\brtx\s+(\d+)\b`), "5090"},
	{regexp.MustCompile(`(?i)\bandroid\s+(\d+)\b`), "16"},
	{regexp.MustCompile(`(?i)\bwindows\s+(\d+)\b`), "11"},
	{regexp.MustCompile(`(?i)\b(\d{2}h[12])\b`), "25h2"},
}

func dateMatchesLabel(plain string, label, date, otherDate *regexp.Regexp) bool {
	for _, line := range strings.Split(plain, "\n") {
		if label.MatchString(line) && date.MatchString(line) && !otherDate.MatchString(line) {
			return true
		}
	}
	return false
}

func assessIntelligenceResponse(text string) (string, []IntelligenceCheck) {
	plain := strings.NewReplacer("*", "", "`", "", "_", "", "|", " ", "\r", "", "’", "'").Replace(text)
	checks := make([]IntelligenceCheck, 0, len(intelligenceReference))
	matched := !intelligenceAmbiguous.MatchString(plain) && !intelligenceNegated.MatchString(plain)
	for _, claim := range intelligenceVersionClaims {
		for _, hit := range claim.pattern.FindAllStringSubmatch(plain, -1) {
			if strings.ToLower(hit[1]) != claim.expected {
				matched = false
			}
		}
	}
	for _, ref := range intelligenceReference {
		ok := ref.pattern.MatchString(plain)
		if ref.item == "iphone" {
			ok = ok && intelligencePro.MatchString(plain) && intelligenceProMax.MatchString(plain)
		}
		if ref.item == "announcement" {
			ok = dateMatchesLabel(plain, intelligenceAnnouncement, ref.pattern, intelligenceReference[2].pattern)
		}
		if ref.item == "on_sale" {
			ok = dateMatchesLabel(plain, intelligenceOnSale, ref.pattern, intelligenceReference[1].pattern)
		}
		checks = append(checks, IntelligenceCheck{Item: ref.item, Expected: ref.expected, Matched: ok})
		matched = matched && ok
	}
	if matched {
		return "passed", checks
	}
	return "manual_review", checks
}

func isIntelligenceTest(ctx context.Context) bool {
	value, _ := ctx.Value(intelligenceTestContextKey{}).(bool)
	return value
}

func createIntelligenceTestPayload() map[string]any {
	payload := createOpenAITestPayload(IntelligenceTestModel, true)
	payload["input"] = []map[string]any{{"type": "message", "role": "user", "content": []map[string]any{{"type": "input_text", "text": IntelligenceTestPrompt}}}}
	payload["instructions"] = "Answer the user's question using only existing knowledge. Do not browse or use tools."
	payload["tools"] = []any{}
	payload["tool_choice"] = "none"
	return payload
}

// Reuses the normal account proxy, OAuth identity and upstream transport, but
// isolates model/prompt from mapping and quota-overdraft prompt injections.
func (s *AccountTestService) TestAccountIntelligence(ctx context.Context, accountID int64) (*IntelligenceTestResult, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth || account.IsSyntheticUITest() {
		return nil, ErrIntelligenceAccountType
	}
	if _, busy := intelligenceTestsRunning.LoadOrStore(accountID, true); busy {
		return nil, ErrIntelligenceTestBusy
	}
	defer intelligenceTestsRunning.Delete(accountID)
	ctx, cancel := context.WithTimeout(ctx, 150*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, intelligenceTestContextKey{}, true)
	started := time.Now()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/", nil).WithContext(ctx)
	testErr := s.testOpenAIAccountConnection(c, account, IntelligenceTestModel, IntelligenceTestPrompt, AccountTestModeDefault)
	text, upstreamError, usage := parseTestSSEOutputWithUsage(recorder.Body.String())
	status, checks := assessIntelligenceResponse(text)
	if testErr != nil || upstreamError != "" || strings.TrimSpace(text) == "" {
		status = "error"
		if upstreamError == "" && testErr != nil {
			upstreamError = testErr.Error()
		}
		if upstreamError == "" {
			upstreamError = "Upstream returned no answer"
		}
	}
	return &IntelligenceTestResult{AccountID: accountID, Model: IntelligenceTestModel, Status: status, ResponseText: text, Error: upstreamError, Checks: checks, TestedAt: started.UTC(), LatencyMS: time.Since(started).Milliseconds(), Usage: usage, Cost: intelligenceTestCost(usage)}, nil
}

// These are the built-in original USD rates for gpt-6-astra. Tests are
// informational only and never use the caller's group multiplier or balance.
func intelligenceTestCost(usage *TestUsage) *IntelligenceTestCost {
	if usage == nil {
		return nil
	}
	const inputRate = 10e-6
	const cachedInputRate = 1e-6
	const outputRate = 50e-6
	cached := usage.CachedInputTokens
	if cached > usage.InputTokens {
		cached = usage.InputTokens
	}
	miss := usage.InputTokens - cached
	inputCost := float64(miss) * inputRate
	cachedCost := float64(cached) * cachedInputRate
	outputCost := float64(usage.OutputTokens) * outputRate
	return &IntelligenceTestCost{
		InputCostUSD:       inputCost,
		CachedInputCostUSD: cachedCost,
		OutputCostUSD:      outputCost,
		TotalCostUSD:       inputCost + cachedCost + outputCost,
		PricingNote:        "gpt-6-astra 内置原价（USD/token），未套用分组倍率；缓存命中按上游 usage 计算",
	}
}
