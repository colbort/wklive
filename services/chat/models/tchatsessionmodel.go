package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatSessionModel = (*customTChatSessionModel)(nil)

const (
	chatSessionStatusWaiting      = 1
	chatSessionStatusPendingAgent = 4
)

type (
	ChatSessionPageFilter struct {
		MerchantId int64
		UserId     int64
		AgentId    int64
		GroupId    int64
		Status     int64
		Priority   int64
		Category   string
		StartTime  int64
		EndTime    int64
	}

	// TChatSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatSessionModel.
	TChatSessionModel interface {
		tChatSessionModel
		FindPage(ctx context.Context, filter ChatSessionPageFilter, cursor int64, limit int64) ([]*TChatSession, int64, error)
		FindOpenByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error)
	}

	customTChatSessionModel struct {
		*defaultTChatSessionModel
	}
)

// NewTChatSessionModel returns a model for the database table.
func NewTChatSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatSessionModel {
	return &customTChatSessionModel{
		defaultTChatSessionModel: newTChatSessionModel(conn, c, opts...),
	}
}

func (m *customTChatSessionModel) FindPage(ctx context.Context, filter ChatSessionPageFilter, cursor int64, limit int64) ([]*TChatSession, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("group_id", filter.GroupId)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("priority", filter.Priority)
	builder.EqString("category", filter.Category)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)
	if filter.AgentId > 0 {
		builder.And("(agent_id = ? OR (agent_id = 0 AND status IN (?, ?)))", filter.AgentId, chatSessionStatusWaiting, chatSessionStatusPendingAgent)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY last_message_time DESC,id DESC LIMIT ?,?", tChatSessionRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatSession
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatSessionModel) FindOpenByUser(ctx context.Context, merchantId int64, userId int64) (*TChatSession, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE merchant_id = ? AND user_id = ? AND status <> ? ORDER BY id DESC LIMIT 1",
		tChatSessionRows,
		m.table,
	)
	var resp TChatSession
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, merchantId, userId, 5); err != nil {
		return nil, err
	}
	return &resp, nil
}
