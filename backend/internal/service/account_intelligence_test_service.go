package service

import (
	"context"
	"encoding/json"
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

// These prompts mirror the v2.8.11 account intelligence tests. The legacy
// knowledge prompt remains available for existing callers and stored history.
const IntelligenceCandyPrompt = `在一个黑色的袋子里放有三种口味的糖果，每种糖果有两种不同的形状（圆形和五角星形，不同的形状靠手感可以分辨）。现已知不同口味的糖和不同形状的数量统计如下表。参赛者需要在活动前决定摸出的糖果数目，那么，最少取出多少个糖果才能保证手中同时拥有不同形状的苹果味和桃子味的糖？（同时手中有圆形苹果味匹配五角星桃子味糖果，或者有圆形桃子味匹配五角星苹果味糖果都满足要求）
苹果味 桃子味 西瓜味
圆形 7 9 8
五角星形 7 6 4`

const IntelligencePelicanPrompt = `创建一个 HTML，内容是 SVG 绘制一个鹈鹕骑自行车的 2D 动画，你不需要任何测试，不要有任何限制`

const IntelligenceQuestionCandy = "candy"
const IntelligenceQuestionPelican = "pelican"
const IntelligenceQuestionKnowledge = "knowledge"

var ErrIntelligenceAccountType = errors.New("intelligence tests require an OpenAI OAuth account")
var ErrIntelligenceTestBusy = errors.New("a test is already running for this account")
var intelligenceTestsRunning sync.Map

const intelligenceTestHistoryExtraKey = "intelligence_test_history"
const intelligenceTestHistoryLimit = 20

type intelligenceTestContextKey struct{}

type intelligenceTestOptions struct {
	questionKind string
	prompt       string
	reasoning    string
}

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
	QuestionKind       string                `json:"question_kind,omitempty"`
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
		// Markdown tables and compact answers sometimes put both dates on one
		// row. The label closest to the expected date is sufficient; requiring
		// the other date to be absent incorrectly rejected those answers.
		labelLoc := label.FindStringIndex(line)
		if labelLoc == nil {
			continue
		}
		dateLoc := date.FindStringIndex(line)
		if dateLoc == nil {
			continue
		}
		otherLoc := otherDate.FindStringIndex(line)
		if otherLoc == nil {
			return true
		}
		// If both dates share a compact table row, associate the label with
		// the nearest date instead of rejecting the entire row.
		labelPos := labelLoc[0]
		dateDistance := absInt(dateLoc[0] - labelPos)
		otherDistance := absInt(otherLoc[0] - labelPos)
		if dateDistance <= otherDistance {
			return true
		}
	}
	return false
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func assessIntelligenceResponse(text string) (string, []IntelligenceCheck) {
	plain := strings.NewReplacer("*", "", "`", "", "_", "", "|", " ", "\r", "", "’", "'").Replace(text)
	checks := make([]IntelligenceCheck, 0, len(intelligenceReference))
	matched := !intelligenceNegated.MatchString(plain)
	for _, line := range strings.Split(plain, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		// A model may repeat the prompt's instruction about uncertainty. That
		// boilerplate is not an uncertain answer; a standalone caveat is.
		if intelligenceAmbiguous.MatchString(trimmed) &&
			!strings.Contains(lower, "if you are uncertain") &&
			!strings.Contains(lower, "if uncertain") &&
			!strings.Contains(lower, "do not browse") &&
			!strings.Contains(lower, "do not speculate") {
			matched = false
		}
	}
	for _, claim := range intelligenceVersionClaims {
		for _, hit := range claim.pattern.FindAllStringSubmatch(plain, -1) {
			if strings.ToLower(hit[1]) != claim.expected {
				matched = false
			}
		}
	}
	for _, ref := range intelligenceReference {
		relevant := intelligenceRelevantText(plain, ref.item)
		ok := ref.pattern.MatchString(relevant)
		if ref.item == "iphone" {
			ok = ok && intelligencePro.MatchString(relevant) && intelligenceProMax.MatchString(relevant)
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

func intelligenceRelevantText(plain, item string) string {
	keywords := map[string][]string{
		"iphone": {"iphone"}, "announcement": {"announcement", "announced", "发布"},
		"on_sale": {"on-sale", "on sale", "发售", "上市"}, "nvidia": {"nvidia", "gpu", "rtx"},
		"android": {"android"}, "macos": {"macos", "mac os"}, "windows": {"windows"},
	}
	var lines []string
	for _, line := range strings.Split(plain, "\n") {
		lower := strings.ToLower(line)
		for _, keyword := range keywords[item] {
			if strings.Contains(lower, strings.ToLower(keyword)) {
				lines = append(lines, line)
				break
			}
		}
	}
	if len(lines) == 0 {
		return plain
	}
	return strings.Join(lines, "\n")
}

func isIntelligenceTest(ctx context.Context) bool {
	if value, ok := ctx.Value(intelligenceTestContextKey{}).(bool); ok {
		return value
	}
	_, ok := ctx.Value(intelligenceTestContextKey{}).(intelligenceTestOptions)
	return ok
}

func intelligenceOptionsFromContext(ctx context.Context) intelligenceTestOptions {
	if value, ok := ctx.Value(intelligenceTestContextKey{}).(intelligenceTestOptions); ok {
		return value
	}
	return intelligenceTestOptions{questionKind: "legacy", prompt: IntelligenceTestPrompt}
}

func createIntelligenceTestPayload(options intelligenceTestOptions) map[string]any {
	prompt := strings.TrimSpace(options.prompt)
	if prompt == "" {
		prompt = IntelligenceTestPrompt
	}
	payload := createOpenAITestPayload(IntelligenceTestModel, true)
	payload["input"] = []map[string]any{{"type": "message", "role": "user", "content": []map[string]any{{"type": "input_text", "text": prompt}}}}
	payload["instructions"] = "Answer the user's question using only existing knowledge. Do not browse or use tools."
	payload["tools"] = []any{}
	payload["tool_choice"] = "none"
	if effort := normalizeIntelligenceReasoning(options.reasoning); effort != "" {
		payload["reasoning"] = map[string]any{"effort": effort}
	}
	return payload
}

func normalizeIntelligenceReasoning(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "minimal", "low", "medium", "high", "xhigh":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func intelligenceQuestion(kind string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case IntelligenceQuestionCandy:
		return IntelligenceQuestionCandy, IntelligenceCandyPrompt
	case IntelligenceQuestionPelican:
		return IntelligenceQuestionPelican, IntelligencePelicanPrompt
	case IntelligenceQuestionKnowledge, "legacy", "":
		return IntelligenceQuestionKnowledge, IntelligenceTestPrompt
	default:
		return IntelligenceQuestionKnowledge, IntelligenceTestPrompt
	}
}

func assessCandyResponse(text string) (string, []IntelligenceCheck) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(text, "，", ","))
	answer := strings.TrimSpace(strings.Trim(trimmed, "。.!！"))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "答案是"))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "答案"))
	answer = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(answer), "answer is"))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "answer"))
	check := IntelligenceCheck{Item: "candy", Expected: "21", Matched: answer == "21"}
	if check.Matched {
		return "passed", []IntelligenceCheck{check}
	}
	return "manual_review", []IntelligenceCheck{check}
}

func assessPelicanResponse(text string) (string, []IntelligenceCheck) {
	plain := strings.ToLower(strings.TrimSpace(text))
	matched := strings.Contains(plain, "<svg")
	check := IntelligenceCheck{Item: "pelican", Expected: "standalone HTML containing SVG", Matched: matched}
	if matched {
		return "passed", []IntelligenceCheck{check}
	}
	return "manual_review", []IntelligenceCheck{check}
}

func assessIntelligenceResponseForQuestion(kind, text string) (string, []IntelligenceCheck) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case IntelligenceQuestionCandy:
		return assessCandyResponse(text)
	case IntelligenceQuestionPelican:
		return assessPelicanResponse(text)
	default:
		return assessIntelligenceResponse(text)
	}
}

// Reuses the normal account proxy, OAuth identity and upstream transport, but
// isolates model/prompt from mapping and quota-overdraft prompt injections.
func (s *AccountTestService) TestAccountIntelligence(ctx context.Context, accountID int64, requested ...string) (*IntelligenceTestResult, error) {
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
	questionKind, prompt := intelligenceQuestion("")
	if len(requested) > 0 {
		questionKind, prompt = intelligenceQuestion(requested[0])
	}
	options := intelligenceTestOptions{questionKind: questionKind, prompt: prompt}
	if len(requested) > 1 {
		options.reasoning = requested[1]
	}
	ctx = context.WithValue(ctx, intelligenceTestContextKey{}, options)
	started := time.Now()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/", nil).WithContext(ctx)
	testErr := s.testOpenAIAccountConnection(c, account, IntelligenceTestModel, prompt, AccountTestModeDefault)
	text, upstreamError, usage := parseTestSSEOutputWithUsage(recorder.Body.String())
	status, checks := assessIntelligenceResponseForQuestion(questionKind, text)
	if testErr != nil || upstreamError != "" || strings.TrimSpace(text) == "" {
		status = "error"
		if upstreamError == "" && testErr != nil {
			upstreamError = testErr.Error()
		}
		if upstreamError == "" {
			upstreamError = "Upstream returned no answer"
		}
	}
	return &IntelligenceTestResult{AccountID: accountID, Model: IntelligenceTestModel, Status: status, ResponseText: text, Error: upstreamError, Checks: checks, QuestionKind: questionKind, TestedAt: started.UTC(), LatencyMS: time.Since(started).Milliseconds(), Usage: usage, Cost: intelligenceTestCost(usage)}, nil
}

// SaveIntelligenceTestResult persists a bounded server-side history in the
// account's extra JSON. Every administrator sees the same recent results.
func (s *AccountTestService) SaveIntelligenceTestResult(ctx context.Context, result *IntelligenceTestResult) error {
	if s == nil || s.accountRepo == nil || result == nil {
		return errors.New("intelligence test persistence is unavailable")
	}
	account, err := s.accountRepo.GetByID(ctx, result.AccountID)
	if err != nil {
		return err
	}
	history := make([]IntelligenceTestResult, 0, intelligenceTestHistoryLimit)
	if raw, ok := account.Extra[intelligenceTestHistoryExtraKey]; ok {
		if encoded, marshalErr := json.Marshal(raw); marshalErr == nil {
			_ = json.Unmarshal(encoded, &history)
		}
	}
	if len(history) >= intelligenceTestHistoryLimit {
		history = history[:intelligenceTestHistoryLimit-1]
	}
	copyResult := *result
	if len(copyResult.ResponseText) > 64*1024 {
		copyResult.ResponseText = copyResult.ResponseText[:64*1024]
	}
	history = append([]IntelligenceTestResult{copyResult}, history...)
	return s.accountRepo.UpdateExtra(ctx, result.AccountID, map[string]any{intelligenceTestHistoryExtraKey: history})
}

func (s *AccountTestService) GetIntelligenceTestHistory(ctx context.Context, accountID int64) ([]IntelligenceTestResult, error) {
	if s == nil || s.accountRepo == nil {
		return nil, errors.New("intelligence test history is unavailable")
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	history := make([]IntelligenceTestResult, 0)
	if raw, ok := account.Extra[intelligenceTestHistoryExtraKey]; ok {
		if encoded, marshalErr := json.Marshal(raw); marshalErr == nil {
			if unmarshalErr := json.Unmarshal(encoded, &history); unmarshalErr != nil {
				return nil, unmarshalErr
			}
		}
	}
	if len(history) > intelligenceTestHistoryLimit {
		history = history[:intelligenceTestHistoryLimit]
	}
	return history, nil
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
