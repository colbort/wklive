package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeOrderPageFilter struct {
	TenantId    int64
	UserId      int64
	ProductId   int64
	OrderNo     string
	ProductName string
	CoinSymbol  string
	Status      int64
	RedeemType  int64
	Source      int64
	StartBegin  int64
	StartEnd    int64
	EndBegin    int64
	EndEnd      int64
}

type StakeOrderModel interface {
	tStakeOrderModel
	FindPage(ctx context.Context, filter StakeOrderPageFilter, cursor int64, limit int64) ([]*TStakeOrder, int64, error)
	SumStakeAmountByStatuses(ctx context.Context, tenantID, user_id, productID int64, statuses []int64) (float64, error)
}

func (m *defaultTStakeOrderModel) FindPage(ctx context.Context, filter StakeOrderPageFilter, cursor int64, limit int64) ([]*TStakeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.TenantId > 0 {
		builder.And("tenant_id = ?", filter.TenantId)
	}
	if filter.UserId > 0 {
		builder.And("user_id = ?", filter.UserId)
	}
	if filter.ProductId > 0 {
		builder.And("product_id = ?", filter.ProductId)
	}
	builder.EqString("order_no", filter.OrderNo)
	if filter.ProductName != "" {
		builder.LikeString("product_name", filter.ProductName)
	}
	builder.EqString("coin_symbol", filter.CoinSymbol)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("redeem_type", filter.RedeemType)
	builder.EqInt64("source", filter.Source)
	builder.GteInt64("start_times", filter.StartBegin)
	builder.LteInt64("start_times", filter.StartEnd)
	builder.GteInt64("end_times", filter.EndBegin)
	builder.LteInt64("end_times", filter.EndEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTStakeOrderModel) SumStakeAmountByStatuses(ctx context.Context, tenantID, user_id, productID int64, statuses []int64) (float64, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	builder.And("user_id = ?", user_id)
	builder.And("product_id = ?", productID)
	builder.InInt64("status", statuses)

	var total sql.NullFloat64
	query := fmt.Sprintf("SELECT COALESCE(SUM(stake_amount), 0) FROM %s WHERE %s", m.table, builder.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, builder.Args()...); err != nil {
		return 0, err
	}
	return total.Float64, nil
}
