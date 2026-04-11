package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type ContractPositionHistoryPageFilter struct {
	TenantId   int64
	UserId     int64
	SymbolId   int64
	MarketType int64
	PositionId int64
	ActionType int64
	TimeStart  int64
	TimeEnd    int64
}

type ContractPositionHistoryModel interface {
	tContractPositionHistoryModel
	FindPage(ctx context.Context, filter ContractPositionHistoryPageFilter, cursor int64, limit int64) ([]*TContractPositionHistory, int64, error)
}

func (m *defaultTContractPositionHistoryModel) FindPage(ctx context.Context, filter ContractPositionHistoryPageFilter, cursor int64, limit int64) ([]*TContractPositionHistory, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("position_id", filter.PositionId)
	builder.EqInt64("action_type", filter.ActionType)
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
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractPositionHistoryRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TContractPositionHistory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
