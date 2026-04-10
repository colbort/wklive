package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type BizTradeEventModel interface {
	tBizTradeEventModel
	FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TBizTradeEvent, int64, error)
}

func (m *defaultTBizTradeEventModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TBizTradeEvent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)

	where := builder.Where()
	args := builder.Args()
	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
			tBizTradeEventRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tBizTradeEventRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TBizTradeEvent
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
