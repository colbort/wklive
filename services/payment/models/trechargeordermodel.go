package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TRechargeOrderModel = (*customTRechargeOrderModel)(nil)

type (
	RechargeOrderPageFilter struct {
		TenantId     int64
		UserId       int64
		OrderNo      string
		Status       int64
		RechargeType int64
	}

	// TRechargeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRechargeOrderModel.
	TRechargeOrderModel interface {
		tRechargeOrderModel
		FindPage(ctx context.Context, filter RechargeOrderPageFilter, cursor int64, limit int64) ([]*TRechargeOrder, int64, error)
	}

	customTRechargeOrderModel struct {
		*defaultTRechargeOrderModel
	}
)

// NewTRechargeOrderModel returns a model for the database table.
func NewTRechargeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRechargeOrderModel {
	return &customTRechargeOrderModel{
		defaultTRechargeOrderModel: newTRechargeOrderModel(conn, c, opts...),
	}
}

func (m *defaultTRechargeOrderModel) FindPage(ctx context.Context, filter RechargeOrderPageFilter, cursor int64, limit int64) ([]*TRechargeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("recharge_type", filter.RechargeType)

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
