package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionOrderPageFilter struct {
	TenantId         int64
	Uid              int64
	AccountId        int64
	ContractId       int64
	UnderlyingSymbol string
	OrderNo          string
	Side             int64
	PositionEffect   int64
	OrderType        int64
	Status           int64
	Statuses         []int64
	CreateTimeStart  int64
	CreateTimeEnd    int64
}

type OptionOrderModel interface {
	tOptionOrderModel
	FindPage(ctx context.Context, filter OptionOrderPageFilter, cursor int64, limit int64) ([]*TOptionOrder, int64, error)
}

func (m *defaultTOptionOrderModel) FindPage(ctx context.Context, filter OptionOrderPageFilter, cursor int64, limit int64) ([]*TOptionOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("uid", filter.Uid)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("position_effect", filter.PositionEffect)
	builder.EqInt64("order_type", filter.OrderType)
	builder.EqInt64("status", filter.Status)
	builder.InInt64("status", filter.Statuses)
	builder.GteInt64("create_times", filter.CreateTimeStart)
	builder.LteInt64("create_times", filter.CreateTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionOrderRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
