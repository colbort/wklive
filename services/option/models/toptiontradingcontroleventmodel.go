package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradingControlEventModel = (*customTOptionTradingControlEventModel)(nil)

type (
	OptionTradingControlEventPageFilter struct {
		TenantId   int64
		UserId     int64
		ContractId int64
		EventType  string
		Reason     string
	}

	// TOptionTradingControlEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradingControlEventModel.
	TOptionTradingControlEventModel interface {
		tOptionTradingControlEventModel
		FindPage(ctx context.Context, filter OptionTradingControlEventPageFilter, cursor, limit int64) ([]*TOptionTradingControlEvent, int64, error)
	}

	customTOptionTradingControlEventModel struct {
		*defaultTOptionTradingControlEventModel
	}
)

func (m *defaultTOptionTradingControlEventModel) FindPage(
	ctx context.Context, filter OptionTradingControlEventPageFilter, cursor, limit int64,
) ([]*TOptionTradingControlEvent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("event_type", filter.EventType)
	builder.EqString("reason", filter.Reason)
	where := builder.Where()
	args := builder.Args()
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionTradingControlEventRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionTradingControlEvent
	if err := m.QueryRowsNoCacheCtx(ctx, &items, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionTradingControlEventModel returns a model for the database table.
func NewTOptionTradingControlEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradingControlEventModel {
	return &customTOptionTradingControlEventModel{
		defaultTOptionTradingControlEventModel: newTOptionTradingControlEventModel(conn, c, opts...),
	}
}

// InsertOptionTradingControlEvent appends an immutable audit fact through the
// caller's SQL connection. Trading admission calls this inside their business
// transaction so audit durability does not depend on Redis availability.
func InsertOptionTradingControlEvent(
	ctx context.Context, conn sqlx.SqlConn, item *TOptionTradingControlEvent,
) error {
	_, err := conn.ExecCtx(ctx, `INSERT INTO t_option_trading_control_event
(tenant_id,user_id,contract_id,order_id,event_type,reason,detail,operator_id,create_times)
VALUES(?,?,?,?,?,?,?,?,?)`,
		item.TenantId, item.UserId, item.ContractId, item.OrderId, item.EventType,
		item.Reason, item.Detail, item.OperatorId, item.CreateTimes,
	)
	return err
}
