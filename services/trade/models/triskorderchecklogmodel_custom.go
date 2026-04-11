package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type RiskOrderCheckLogPageFilter struct {
	TenantId    int64
	UserId      int64
	SymbolId    int64
	MarketType  int64
	CheckType   int64
	CheckResult int64
	TimeStart   int64
	TimeEnd     int64
}

type RiskOrderCheckLogModel interface {
	tRiskOrderCheckLogModel
	FindPage(ctx context.Context, filter RiskOrderCheckLogPageFilter, cursor int64, limit int64) ([]*TRiskOrderCheckLog, int64, error)
}

func (m *defaultTRiskOrderCheckLogModel) FindPage(ctx context.Context, filter RiskOrderCheckLogPageFilter, cursor int64, limit int64) ([]*TRiskOrderCheckLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("check_type", filter.CheckType)
	builder.EqInt64("check_result", filter.CheckResult)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tRiskOrderCheckLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TRiskOrderCheckLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
