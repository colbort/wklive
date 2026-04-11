package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeOrderModel interface {
	tStakeOrderModel
	FindPage(ctx context.Context, tenantID int64, cursor, limit int64, uid, productID int64, orderNo, productName, coinSymbol string, status, redeemType, source int64, startBegin, startEnd, endBegin, endEnd int64) ([]*TStakeOrder, int64, error)
	SumStakeAmountByStatuses(ctx context.Context, tenantID, uid, productID int64, statuses []int64) (float64, error)
}

func (m *defaultTStakeOrderModel) FindPage(ctx context.Context, tenantID int64, cursor, limit int64, uid, productID int64, orderNo, productName, coinSymbol string, status, redeemType, source int64, startBegin, startEnd, endBegin, endEnd int64) ([]*TStakeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	if uid > 0 {
		builder.And("uid = ?", uid)
	}
	if productID > 0 {
		builder.And("product_id = ?", productID)
	}
	builder.EqString("order_no", orderNo)
	if productName != "" {
		builder.LikeString("product_name", "%"+productName+"%")
	}
	builder.EqString("coin_symbol", coinSymbol)
	builder.EqInt64("status", status)
	builder.EqInt64("redeem_type", redeemType)
	builder.EqInt64("source", source)
	builder.GteInt64("start_times", startBegin)
	builder.LteInt64("start_times", startEnd)
	builder.GteInt64("end_times", endBegin)
	builder.LteInt64("end_times", endEnd)

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

func (m *defaultTStakeOrderModel) SumStakeAmountByStatuses(ctx context.Context, tenantID, uid, productID int64, statuses []int64) (float64, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	builder.And("uid = ?", uid)
	builder.And("product_id = ?", productID)
	builder.InInt64("status", statuses)

	var total sql.NullFloat64
	query := fmt.Sprintf("SELECT COALESCE(SUM(stake_amount), 0) FROM %s WHERE %s", m.table, builder.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, builder.Args()...); err != nil {
		return 0, err
	}
	return total.Float64, nil
}
