package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatAssignmentModel = (*customTChatAssignmentModel)(nil)

type (
	ChatAssignmentPageFilter struct {
		MerchantId  int64
		SessionNo   string
		FromAgentId int64
		ToAgentId   int64
		AssignType  int64
		StartTime   int64
		EndTime     int64
	}

	// TChatAssignmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatAssignmentModel.
	TChatAssignmentModel interface {
		tChatAssignmentModel
		FindPage(ctx context.Context, filter ChatAssignmentPageFilter, cursor int64, limit int64) ([]*TChatAssignment, int64, error)
		ListBySessionNo(ctx context.Context, sessionNo string) ([]*TChatAssignment, error)
	}

	customTChatAssignmentModel struct {
		*defaultTChatAssignmentModel
	}
)

// NewTChatAssignmentModel returns a model for the database table.
func NewTChatAssignmentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatAssignmentModel {
	return &customTChatAssignmentModel{
		defaultTChatAssignmentModel: newTChatAssignmentModel(conn, c, opts...),
	}
}

func (m *customTChatAssignmentModel) FindPage(ctx context.Context, filter ChatAssignmentPageFilter, cursor int64, limit int64) ([]*TChatAssignment, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqString("session_no", filter.SessionNo)
	builder.EqInt64("from_agent_id", filter.FromAgentId)
	builder.EqInt64("to_agent_id", filter.ToAgentId)
	builder.EqInt64("assign_type", filter.AssignType)
	builder.GteInt64("create_times", filter.StartTime)
	builder.LteInt64("create_times", filter.EndTime)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", tChatAssignmentRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatAssignment
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatAssignmentModel) ListBySessionNo(ctx context.Context, sessionNo string) ([]*TChatAssignment, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE session_no = ? ORDER BY id ASC", tChatAssignmentRows, m.table)
	var list []*TChatAssignment
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, sessionNo); err != nil {
		return nil, err
	}
	return list, nil
}
