package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// ListModelPlazaAccounts reads only model configuration and active group
// bindings. Authentication secrets and scheduler state are not selected.
func (r *accountRepository) ListModelPlazaAccounts(ctx context.Context) ([]service.Account, error) {
	if r.sql == nil {
		return nil, fmt.Errorf("model catalog SQL executor not configured")
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT a.id, a.platform, a.type,
			jsonb_build_object('model_mapping', a.credentials->'model_mapping', 'oauth_type', a.credentials->'oauth_type'),
			jsonb_build_object('openai_passthrough', a.extra->'openai_passthrough', 'openai_oauth_passthrough', a.extra->'openai_oauth_passthrough'),
			jsonb_agg(ag.group_id ORDER BY ag.group_id)
		FROM accounts a
		JOIN account_groups ag ON ag.account_id = a.id
		JOIN groups g ON g.id = ag.group_id AND g.deleted_at IS NULL AND g.status = 'active'
		WHERE a.deleted_at IS NULL AND a.status = 'active'
		GROUP BY a.id
		ORDER BY a.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []service.Account
	for rows.Next() {
		var account service.Account
		var credentials, extra, groups []byte
		if err := rows.Scan(&account.ID, &account.Platform, &account.Type, &credentials, &extra, &groups); err != nil {
			return nil, err
		}
		for _, field := range []struct {
			data   []byte
			target any
		}{
			{credentials, &account.Credentials}, {extra, &account.Extra}, {groups, &account.GroupIDs},
		} {
			if err := json.Unmarshal(field.data, field.target); err != nil {
				return nil, fmt.Errorf("decode model catalog configuration: %w", err)
			}
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}
