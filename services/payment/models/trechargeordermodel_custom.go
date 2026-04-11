package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type RechargeOrderModel interface {
	tRechargeOrderModel
	FindPage(ctx context.Context, tenantId int64, userId int64, orderNo string, status int64, cursor int64, limit int64) ([]*TRechargeOrder, int64, error)
}

func (m *defaultTRechargeOrderModel) FindPage(ctx context.Context, tenantId int64, userId int64, orderNo string, status int64, cursor int64, limit int64) ([]*TRechargeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("user_id", userId)
	builder.EqString("order_no", orderNo)
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tRechargeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TRechargeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
