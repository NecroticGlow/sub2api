package repository

import (
	"context"
	"encoding/json"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPelicanManualHistoryAtomicAppend(t *testing.T) {
	for _, affected := range []int64{0, 1} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		repo := &accountRepository{sql: db}
		record := &service.PelicanManualRecord{ID: "batch", CreatedAt: "2026-09-25T00:00:00Z", ModelID: "test-model", Runs: []json.RawMessage{json.RawMessage(`{"status":"success","output":"21"}`)}}
		mock.ExpectExec(`(?s)UPDATE accounts SET extra = jsonb_set.*WITH ORDINALITY.*LIMIT 8.*deleted_at IS NULL`).WithArgs(int64(42), sqlmock.AnyArg(), "batch").WillReturnResult(sqlmock.NewResult(0, affected))
		err = repo.AppendPelicanManualRecord(context.Background(), 42, record)
		if affected == 0 {
			require.ErrorIs(t, err, service.ErrAccountNotFound)
		} else {
			require.NoError(t, err)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestPelicanManualHistorySurvivesStaleAccountEdit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	mock.ExpectQuery(`(?s)SELECT.*FOR NO KEY UPDATE`).
		WillReturnRows(sqlmock.NewRows([]string{"identity_unchanged", "ollama_group_unchanged", "ollama_proxy_unchanged", "enabled", "rate_sync_enabled", "snapshot", "ollama_session", "ollama_auto", "ollama_snapshot", "opencode_group_unchanged", "opencode_auto", "opencode_snapshot", "current_extra"}).
			AddRow(false, false, false, nil, nil, nil, nil, nil, nil, false, nil, nil, []byte(`{"pelican_manual_history":[{"id":"latest"}]}`)))
	account := &service.Account{ID: 42, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth}
	extra, err := lockAndMergeAccountProbeExtra(context.Background(), client, account, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []any{map[string]any{"id": "latest"}}, extra[service.PelicanManualHistoryKey])
	require.NoError(t, mock.ExpectationsWereMet())
}
