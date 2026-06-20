package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeCancelLogModel = (*customTTradeCancelLogModel)(nil)

type (
	TradeCancelLogPageFilter struct {
		TenantId     int64
		UserId       int64
		OrderId      int64
		OrderNo      string
		CancelSource int64
		TimeStart    int64
		TimeEnd      int64
	}

	// TTradeCancelLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeCancelLogModel.
	TTradeCancelLogModel interface {
		tTradeCancelLogModel
		FindPage(ctx context.Context, filter TradeCancelLogPageFilter, cursor int64, limit int64) ([]*TTradeCancelLog, int64, error)
	}

	customTTradeCancelLogModel struct {
		*defaultTTradeCancelLogModel
	}
)

// NewTTradeCancelLogModel returns a model for the database table.
func NewTTradeCancelLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeCancelLogModel {
	return &customTTradeCancelLogModel{
		defaultTTradeCancelLogModel: newTTradeCancelLogModel(conn, c, opts...),
	}
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
