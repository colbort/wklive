package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type WithdrawNotifyLogPageFilter struct {
	TenantId        int64
	OrderNo         string
	OrderId         int64
	PlatformId      int64
	ChannelId       int64
	NotifyStatus    int64
	SignResult      int64
	CreateTimeStart int64
	CreateTimeEnd   int64
}

type WithdrawNotifyLogModel interface {
	tWithdrawNotifyLogModel
	FindPage(ctx context.Context, filter WithdrawNotifyLogPageFilter, cursor int64, limit int64) ([]*TWithdrawNotifyLog, int64, error)
}

func (m *defaultTWithdrawNotifyLogModel) FindPage(ctx context.Context, filter WithdrawNotifyLogPageFilter, cursor int64, limit int64) ([]*TWithdrawNotifyLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("order_id", filter.OrderId)
	builder.EqInt64("platform_id", filter.PlatformId)
	builder.EqInt64("channel_id", filter.ChannelId)
	builder.EqInt64("notify_status", filter.NotifyStatus)
	builder.EqInt64("sign_result", filter.SignResult)
	builder.GteInt64("create_times", filter.CreateTimeStart)
	builder.LteInt64("create_times", filter.CreateTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tWithdrawNotifyLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TWithdrawNotifyLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
