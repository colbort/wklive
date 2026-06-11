package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type TradeSymbolLeverageConfigModel interface {
	tTradeSymbolLeverageConfigModel
	FindPage(ctx context.Context, tenantId int64, symbolId int64, marketType int64, marginMode int64, enabled int64, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error)
}

func (m *defaultTTradeSymbolLeverageConfigModel) FindPage(ctx context.Context, tenantId int64, symbolId int64, marketType int64, marginMode int64, enabled int64, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("symbol_id", symbolId)
	builder.EqInt64("market_type", marketType)
	builder.EqInt64("margin_mode", marginMode)
	if enabled > 0 {
		builder.And("enabled = ?", enabled)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string
	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s FROM %s WHERE %s ORDER BY sort ASC, id DESC LIMIT ?`,
			tTradeSymbolLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s FROM %s WHERE %s AND id < ? ORDER BY sort ASC, id DESC LIMIT ?`,
			tTradeSymbolLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TTradeSymbolLeverageConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
