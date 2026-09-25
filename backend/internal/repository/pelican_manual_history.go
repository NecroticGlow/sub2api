package repository

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// Append under the row lock in one UPDATE so parallel administrators cannot
// overwrite each other's history. Other account extra and credentials stay intact.
func (r *accountRepository) AppendPelicanManualRecord(ctx context.Context, id int64, record *service.PelicanManualRecord) error {
	if err := service.ValidatePelicanManualRecord(record); err != nil {
		return err
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	result, err := r.sql.ExecContext(ctx, `UPDATE accounts SET extra = jsonb_set(COALESCE(extra, '{}'::jsonb), '{pelican_manual_history}',
  (SELECT jsonb_agg(value ORDER BY ord) FROM (
    SELECT value, ord FROM jsonb_array_elements(jsonb_build_array($2::jsonb) ||
      CASE WHEN jsonb_typeof(extra->'pelican_manual_history') = 'array' THEN extra->'pelican_manual_history' ELSE '[]'::jsonb END)
    WITH ORDINALITY AS entries(value, ord)
    WHERE ord = 1 OR value->>'id' IS DISTINCT FROM $3
    ORDER BY ord LIMIT 8
  ) recent), true) WHERE id = $1 AND deleted_at IS NULL`, id, string(encoded), record.ID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrAccountNotFound
	}
	return nil
}
