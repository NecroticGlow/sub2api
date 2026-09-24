package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// ListModelPlazaUsage uses the same successful billing-log criterion as usage
// analytics, counting requests rather than tokens or spend. Only public groups
// contribute; individual user/account/request details never leave this query.
func (r *usageLogRepository) ListModelPlazaUsage(ctx context.Context, start, end time.Time) ([]service.PlazaModelUsage, error) {
	modelExpr := "LOWER(" + resolveModelDimensionExpressionWithAlias("requested", "ul") + ")"
	query := `SELECT ul.group_id, ` + usageLogEffectivePlatformExpr + ` AS platform,
		` + modelExpr + ` AS model, COUNT(*)
		FROM usage_logs ul
		JOIN groups g ON g.id = ul.group_id
		LEFT JOIN accounts a ON a.id = ul.account_id
		WHERE ul.created_at >= $1 AND ul.created_at < $2
		AND ` + usageLogSuccessFilterUL + `
		AND g.status = 'active' AND g.deleted_at IS NULL AND NOT g.is_exclusive
		GROUP BY ul.group_id, ` + usageLogEffectivePlatformExpr + `, ` + modelExpr
	rows, err := r.sql.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]service.PlazaModelUsage, 0)
	for rows.Next() {
		var row service.PlazaModelUsage
		if err := rows.Scan(&row.GroupID, &row.Platform, &row.Model, &row.Requests); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
