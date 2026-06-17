package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeRedeemLogPageFilter struct {
	TenantId     int64
	UserId       int64
	OrderId      int64
	ProductId    int64
	OrderNo      string
	RedeemNo     string
	RedeemType   int64
	RedeemStatus int64
	RedeemBegin  int64
	RedeemEnd    int64
}

type StakeRedeemLogModel interface {
	tStakeRedeemLogModel
	FindPage(ctx context.Context, filter StakeRedeemLogPageFilter, cursor int64, limit int64) ([]*TStakeRedeemLog, int64, error)
}

func (m *defaultTStakeRedeemLogModel) FindPage(ctx context.Context, filter StakeRedeemLogPageFilter, cursor int64, limit int64) ([]*TStakeRedeemLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.TenantId > 0 {
		builder.And("tenant_id = ?", filter.TenantId)
	}
	if filter.UserId > 0 {
		builder.And("user_id = ?", filter.UserId)
	}
	if filter.OrderId > 0 {
		builder.And("order_id = ?", filter.OrderId)
	}
	if filter.ProductId > 0 {
		builder.And("product_id = ?", filter.ProductId)
	}
	builder.EqString("order_no", filter.OrderNo)
	builder.EqString("redeem_no", filter.RedeemNo)
	builder.EqInt64("redeem_type", filter.RedeemType)
	builder.EqInt64("redeem_status", filter.RedeemStatus)
	builder.GteInt64("redeem_times", filter.RedeemBegin)
	builder.LteInt64("redeem_times", filter.RedeemEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeRedeemLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeRedeemLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
