package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAISessionCookieRepository struct{ db *sql.DB }

func NewOpenAISessionCookieRepository(db *sql.DB) service.OpenAISessionCookieRepository {
	return &openAISessionCookieRepository{db: db}
}

func (r *openAISessionCookieRepository) Get(ctx context.Context, accountID int64) (string, error) {
	var ciphertext string
	err := r.db.QueryRowContext(ctx, `SELECT cookie_ciphertext FROM openai_session_cookies WHERE account_id = $1`, accountID).Scan(&ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return ciphertext, err
}

func (r *openAISessionCookieRepository) Save(ctx context.Context, accountID int64, ciphertext string) error {
	result, err := r.db.ExecContext(ctx, `WITH target AS (
 SELECT id FROM accounts WHERE id = $1 AND deleted_at IS NULL
 AND platform = 'openai' AND type = 'oauth' AND parent_account_id IS NULL FOR UPDATE
)
INSERT INTO openai_session_cookies (account_id, cookie_ciphertext)
 SELECT id, $2 FROM target
 ON CONFLICT (account_id) DO UPDATE SET cookie_ciphertext = EXCLUDED.cookie_ciphertext, updated_at = NOW()`, accountID, ciphertext)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("account is no longer eligible for Session storage")
	}
	return nil
}

func (r *openAISessionCookieRepository) Delete(ctx context.Context, accountID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM openai_session_cookies WHERE account_id = $1`, accountID)
	return err
}
