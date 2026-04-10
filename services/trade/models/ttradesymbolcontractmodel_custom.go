package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type TradeSymbolContractModel interface {
	tTradeSymbolContractModel
	FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeSymbolContract, int64, error)
}

func (m *defaultTTradeSymbolContractModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeSymbolContract, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
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
			tTradeSymbolContractRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tTradeSymbolContractRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TTradeSymbolContract
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
