package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type TradeCancelLogPageFilter struct {
	TenantId     int64
	UserId       int64
	OrderId      int64
	OrderNo      string
	CancelSource int64
	TimeStart    int64
	TimeEnd      int64
}

type TradeCancelLogModel interface {
	tTradeCancelLogModel
	FindPage(ctx context.Context, filter TradeCancelLogPageFilter, cursor int64, limit int64) ([]*TTradeCancelLog, int64, error)
}

func (m *defaultTTradeCancelLogModel) FindPage(ctx context.Context, filter TradeCancelLogPageFilter, cursor int64, limit int64) ([]*TTradeCancelLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("order_id", filter.OrderId)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("cancel_source", filter.CancelSource)
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
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeCancelLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeCancelLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
