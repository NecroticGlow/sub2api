package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const PelicanManualHistoryKey = "pelican_manual_history"
const PelicanManualHistoryLimit = 8

// Manual records are admin-authored display history, not authoritative billing
// or quality judgments. Scheduled quality actions continue to use server results.
type PelicanManualRecord struct {
	ID              string            `json:"id"`
	CreatedAt       string            `json:"createdAt"`
	QuestionKind    string            `json:"questionKind,omitempty"`
	Prompt          string            `json:"prompt"`
	ModelID         string            `json:"modelId"`
	ReasoningEffort string            `json:"reasoningEffort"`
	Runs            []json.RawMessage `json:"runs"`
}

func ValidatePelicanManualRecord(record *PelicanManualRecord) error {
	if record == nil || len(record.ID) == 0 || len(record.ID) > 100 || strings.TrimSpace(record.ModelID) == "" || len(record.ModelID) > 200 || len(record.Runs) < 1 || len(record.Runs) > 8 {
		return errors.New("invalid manual test record")
	}
	if _, err := time.Parse(time.RFC3339Nano, record.CreatedAt); err != nil {
		return errors.New("invalid record timestamp")
	}
	if record.QuestionKind != "" && record.QuestionKind != "candy" && record.QuestionKind != "pelican" && record.QuestionKind != "knowledge" {
		return errors.New("invalid question kind")
	}
	for _, run := range record.Runs {
		var value struct {
			Status string `json:"status"`
			Output string `json:"output"`
			Error  string `json:"error"`
		}
		if json.Unmarshal(run, &value) != nil || (value.Status != "success" && value.Status != "error") {
			return errors.New("invalid completed run")
		}
	}
	encoded, err := json.Marshal(record)
	if err != nil || len(encoded) > 256<<10 {
		return errors.New("manual test record exceeds 256 KiB")
	}
	return nil
}

func (s *AccountTestService) SavePelicanManualRecord(ctx context.Context, accountID int64, record *PelicanManualRecord) error {
	if err := ValidatePelicanManualRecord(record); err != nil {
		return err
	}
	store, ok := s.accountRepo.(interface {
		AppendPelicanManualRecord(context.Context, int64, *PelicanManualRecord) error
	})
	if !ok {
		return errors.New("manual history storage unavailable")
	}
	return store.AppendPelicanManualRecord(ctx, accountID, record)
}

func (s *AccountTestService) GetPelicanManualHistory(ctx context.Context, accountID int64) ([]PelicanManualRecord, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	items := []PelicanManualRecord{}
	if value, ok := account.Extra[PelicanManualHistoryKey]; ok {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(encoded, &items); err != nil {
			return nil, err
		}
	}
	if len(items) > PelicanManualHistoryLimit {
		items = items[:PelicanManualHistoryLimit]
	}
	return items, nil
}
