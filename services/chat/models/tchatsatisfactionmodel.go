package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatSatisfactionModel = (*customTChatSatisfactionModel)(nil)

type (
	ChatSatisfactionPageFilter struct {
		MerchantId int64
		SessionNo  string
		UserId     int64
		AgentId    int64
		Score      int64
		StartTime  int64
		EndTime    int64
	}

	// TChatSatisfactionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatSatisfactionModel.
	TChatSatisfactionModel interface {
		tChatSatisfactionModel
		FindPage(ctx context.Context, filter ChatSatisfactionPageFilter, cursor int64, limit int64) ([]*TChatSatisfaction, int64, error)
	}

	customTChatSatisfactionModel struct {
		*defaultTChatSatisfactionModel
	}
)

// NewTChatSatisfactionModel returns a model for the database table.
func NewTChatSatisfactionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatSatisfactionModel {
	return &customTChatSatisfactionModel{
		defaultTChatSatisfactionModel: newTChatSatisfactionModel(conn, c, opts...),
	}
}

func (m *customTChatSatisfactionModel) FindPage(ctx context.Context, filter ChatSatisfactionPageFilter, cursor int64, limit int64) ([]*TChatSatisfaction, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqString("session_no", filter.SessionNo)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("agent_id", filter.AgentId)
	builder.EqInt64("score", filter.Score)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatSatisfactionRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatSatisfaction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
