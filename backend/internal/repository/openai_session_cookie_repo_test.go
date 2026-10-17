package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestOpenAISessionCookieRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewOpenAISessionCookieRepository(db)
	ctx := context.Background()
	mock.ExpectQuery("SELECT cookie_ciphertext FROM openai_session_cookies").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"cookie_ciphertext"}))
	value, err := repo.Get(ctx, 7)
	require.NoError(t, err)
	require.Empty(t, value)
	mock.ExpectExec("(?s)WITH target AS.*FOR UPDATE.*INSERT INTO openai_session_cookies").WithArgs(int64(7), "encrypted-test-value").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Save(ctx, 7, "encrypted-test-value"))
	mock.ExpectQuery("SELECT cookie_ciphertext FROM openai_session_cookies").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"cookie_ciphertext"}).AddRow("encrypted-test-value"))
	value, err = repo.Get(ctx, 7)
	require.NoError(t, err)
	require.Equal(t, "encrypted-test-value", value)
	mock.ExpectExec("DELETE FROM openai_session_cookies").WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(ctx, 7))
	mock.ExpectExec("(?s)WITH target AS.*FOR UPDATE.*INSERT INTO openai_session_cookies").WithArgs(int64(8), "encrypted-test-value").WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorContains(t, repo.Save(ctx, 8, "encrypted-test-value"), "no longer eligible")
	require.NoError(t, mock.ExpectationsWereMet())
}
