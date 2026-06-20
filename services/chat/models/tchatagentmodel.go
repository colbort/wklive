package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatAgentModel = (*customTChatAgentModel)(nil)

type (
	ChatAgentPageFilter struct {
		MerchantId int64
		ChatUserId int64
		GroupId    int64
		AgentNo    string
		Status     int64
	}

	// TChatAgentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatAgentModel.
	TChatAgentModel interface {
		tChatAgentModel
		FindPage(ctx context.Context, filter ChatAgentPageFilter, cursor int64, limit int64) ([]*TChatAgent, int64, error)
		FindAvailable(ctx context.Context, merchantId int64, groupId int64, limit int64) ([]*TChatAgent, error)
	}

	customTChatAgentModel struct {
		*defaultTChatAgentModel
	}
)

// NewTChatAgentModel returns a model for the database table.
func NewTChatAgentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatAgentModel {
	return &customTChatAgentModel{
		defaultTChatAgentModel: newTChatAgentModel(conn, c, opts...),
	}
}

func (m *customTChatAgentModel) FindPage(ctx context.Context, filter ChatAgentPageFilter, cursor int64, limit int64) ([]*TChatAgent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("chat_user_id", filter.ChatUserId)
	builder.EqInt64("group_id", filter.GroupId)
	builder.EqString("agent_no", filter.AgentNo)
	builder.EqInt64("status", filter.Status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatAgentRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatAgent
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatAgentModel) FindAvailable(ctx context.Context, merchantId int64, groupId int64, limit int64) ([]*TChatAgent, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", merchantId)
	builder.EqInt64("group_id", groupId)
	builder.And("status = ?", 2)
	builder.And("current_session_count < max_session_count")

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY current_session_count ASC,id ASC LIMIT ?",
		tChatAgentRows,
		m.table,
		builder.Where(),
	)
	args := append(builder.Args(), limit)

	var list []*TChatAgent
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}
