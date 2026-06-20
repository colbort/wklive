package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatWorkOrderModel = (*customTChatWorkOrderModel)(nil)

type (
	ChatWorkOrderPageFilter struct {
		Keyword    string
		MerchantId int64
		SessionNo  string
		UserId     int64
		AgentId    int64
		GroupId    int64
		Priority   int64
		Status     int64
		HandlerId  int64
		StartTime  int64
		EndTime    int64
	}

	// TChatWorkOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatWorkOrderModel.
	TChatWorkOrderModel interface {
		tChatWorkOrderModel
		FindPage(ctx context.Context, filter ChatWorkOrderPageFilter, cursor int64, limit int64) ([]*TChatWorkOrder, int64, error)
	}

	customTChatWorkOrderModel struct {
		*defaultTChatWorkOrderModel
	}
)

// NewTChatWorkOrderModel returns a model for the database table.
func NewTChatWorkOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatWorkOrderModel {
	return &customTChatWorkOrderModel{
		defaultTChatWorkOrderModel: newTChatWorkOrderModel(conn, c, opts...),
	}
}

func (m *customTChatWorkOrderModel) FindPage(ctx context.Context, filter ChatWorkOrderPageFilter, cursor int64, limit int64) ([]*TChatWorkOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(work_order_no LIKE ? OR title LIKE ? OR contact_name LIKE ? OR contact_mobile LIKE ? OR contact_email LIKE ?)", like, like, like, like, like)
	}
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqString("session_no", filter.SessionNo)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("agent_id", filter.AgentId)
	builder.EqInt64("group_id", filter.GroupId)
	builder.EqInt64("priority", filter.Priority)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("handler_id", filter.HandlerId)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatWorkOrderRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatWorkOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
