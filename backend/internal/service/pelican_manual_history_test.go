package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func manualHistoryFixture() *PelicanManualRecord {
	return &PelicanManualRecord{ID: "batch-1", CreatedAt: "2026-09-25T00:00:00Z", QuestionKind: "knowledge", ModelID: "gpt-6-astra", Runs: []json.RawMessage{json.RawMessage(`{"status":"success","output":"uncertain"}`)}}
}

func TestPelicanManualHistoryValidation(t *testing.T) {
	require.NoError(t, ValidatePelicanManualRecord(manualHistoryFixture()))
	for _, change := range []func(*PelicanManualRecord){
		func(r *PelicanManualRecord) { r.ID = "" },
		func(r *PelicanManualRecord) { r.ModelID = " " },
		func(r *PelicanManualRecord) { r.QuestionKind = "unknown" },
		func(r *PelicanManualRecord) { r.CreatedAt = "invalid" },
		func(r *PelicanManualRecord) { r.Runs = nil },
		func(r *PelicanManualRecord) { r.Runs[0] = json.RawMessage(`{"status":"running"}`) },
		func(r *PelicanManualRecord) { r.Prompt = strings.Repeat("x", 256<<10) },
	} {
		record := manualHistoryFixture()
		change(record)
		require.Error(t, ValidatePelicanManualRecord(record))
	}
}

type manualHistoryRepoStub struct {
	AccountRepository
	account *Account
	saved   *PelicanManualRecord
}

func (s *manualHistoryRepoStub) GetByID(context.Context, int64) (*Account, error) {
	return s.account, nil
}
func (s *manualHistoryRepoStub) AppendPelicanManualRecord(_ context.Context, _ int64, record *PelicanManualRecord) error {
	s.saved = record
	return nil
}

func TestPelicanManualHistoryReadWriteWithoutUpstream(t *testing.T) {
	repo := &manualHistoryRepoStub{account: &Account{ID: 42, Extra: map[string]any{}}}
	svc := &AccountTestService{accountRepo: repo}
	items, err := svc.GetPelicanManualHistory(context.Background(), 42)
	require.NoError(t, err)
	require.Empty(t, items)
	record := manualHistoryFixture()
	require.NoError(t, svc.SavePelicanManualRecord(context.Background(), 42, record))
	require.Same(t, record, repo.saved)
	repo.account.Extra[PelicanManualHistoryKey] = []*PelicanManualRecord{record}
	items, err = svc.GetPelicanManualHistory(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, "batch-1", items[0].ID)
}

func TestIntelligenceKnowledgeUsesPlainTextContract(t *testing.T) {
	for _, kind := range []string{"knowledge", "japan_pm"} {
		t.Run(kind, func(t *testing.T) {
			cfg := &PelicanTestConfig{QuestionKind: kind, Prompt: "仅依据已有知识回答。"}
			require.Contains(t, intelligenceTestPrompt(cfg), "不联网、不调用工具、不猜测")
			require.NotContains(t, intelligenceTestPrompt(cfg), PelicanDeliveryContract)
			require.Empty(t, intelligenceTestOutputError(cfg, "uncertain"))
			require.NotEmpty(t, intelligenceTestOutputError(cfg, " "))
			record := manualHistoryFixture()
			record.QuestionKind = kind
			require.NoError(t, ValidatePelicanManualRecord(record))
			plan := pelicanPlan()
			plan.PelicanConfig.QuestionKind = kind
			_, err := nextPlanRun(plan, time.Now())
			require.NoError(t, err)
			plan.PelicanConfig.TestChannel = "bps"
			_, err = nextPlanRun(plan, time.Now())
			require.Error(t, err, "knowledge questions cannot use candy-only BPS channel")
		})
	}
}
