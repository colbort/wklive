package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeSymbolLeverageConfigModel = (*customTTradeSymbolLeverageConfigModel)(nil)

type (
	TradeSymbolLeverageConfigPageFilter struct {
		TenantId   int64
		SymbolId   int64
		MarketType int64
		MarginMode int64
		Enabled    int64
	}

	TTradeSymbolLeverageConfigModel interface {
		tTradeSymbolLeverageConfigModel
		FindPage(ctx context.Context, filter TradeSymbolLeverageConfigPageFilter, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error)
	}

	customTTradeSymbolLeverageConfigModel struct {
		*defaultTTradeSymbolLeverageConfigModel
	}
)

func NewTTradeSymbolLeverageConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolLeverageConfigModel {
	return &customTTradeSymbolLeverageConfigModel{
		defaultTTradeSymbolLeverageConfigModel: newTTradeSymbolLeverageConfigModel(conn, c, opts...),
	}
}

func (m *defaultTTradeSymbolLeverageConfigModel) FindPage(ctx context.Context, filter TradeSymbolLeverageConfigPageFilter, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("margin_mode", filter.MarginMode)
	if filter.Enabled > 0 {
		builder.And("enabled = ?", filter.Enabled)
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
