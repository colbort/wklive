package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RechargeNotifyLogModel interface {
	tRechargeNotifyLogModel
	FindPage(ctx context.Context, tenantId int64, orderNo string, orderId int64, platformId int64, channelId int64, notifyStatus int64, signResult int64, createTimeStart int64, createTimeEnd int64, cursor int64, limit int64) ([]*TRechargeNotifyLog, int64, error)
	FindOneByOrderId(ctx context.Context, orderId int64) (*TRechargeNotifyLog, error)
}

func (m *defaultTRechargeNotifyLogModel) FindPage(ctx context.Context, tenantId int64, orderNo string, orderId int64, platformId int64, channelId int64, notifyStatus int64, signResult int64, createTimeStart int64, createTimeEnd int64, cursor int64, limit int64) ([]*TRechargeNotifyLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqString("order_no", orderNo)
	builder.EqInt64("order_id", orderId)
	builder.EqInt64("platform_id", platformId)
	builder.EqInt64("channel_id", channelId)
	builder.EqInt64("notify_status", notifyStatus)
	builder.EqInt64("sign_result", signResult)
	builder.GteInt64("create_times", createTimeStart)
	builder.LteInt64("create_times", createTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tRechargeNotifyLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TRechargeNotifyLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTRechargeNotifyLogModel) FindOneByOrderId(ctx context.Context, orderId int64) (*TRechargeNotifyLog, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE order_id = ? ORDER BY id DESC LIMIT 1", tRechargeNotifyLogRows, m.table)

	var resp TRechargeNotifyLog
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, orderId); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}
