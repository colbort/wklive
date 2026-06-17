package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type WithdrawOrderPageFilter struct {
	TenantId int64
	UserId   int64
	OrderNo  string
	Status   int64
}

type WithdrawOrderModel interface {
	tWithdrawOrderModel
	FindPage(ctx context.Context, filter WithdrawOrderPageFilter, cursor int64, limit int64) ([]*TWithdrawOrder, int64, error)
}

func (m *defaultTWithdrawOrderModel) FindPage(ctx context.Context, filter WithdrawOrderPageFilter, cursor int64, limit int64) ([]*TWithdrawOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("status", filter.Status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tWithdrawOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TWithdrawOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
