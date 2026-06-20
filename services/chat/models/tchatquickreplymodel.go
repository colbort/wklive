package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatQuickReplyModel = (*customTChatQuickReplyModel)(nil)

type (
	ChatQuickReplyPageFilter struct {
		Keyword    string
		MerchantId int64
		AgentId    int64
		CategoryId int64
		Enabled    int64
	}

	// TChatQuickReplyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatQuickReplyModel.
	TChatQuickReplyModel interface {
		tChatQuickReplyModel
		FindPage(ctx context.Context, filter ChatQuickReplyPageFilter, cursor int64, limit int64) ([]*TChatQuickReply, int64, error)
		ListEnabled(ctx context.Context, merchantId int64, agentId int64, categoryId int64) ([]*TChatQuickReply, error)
	}

	customTChatQuickReplyModel struct {
		*defaultTChatQuickReplyModel
	}
)

// NewTChatQuickReplyModel returns a model for the database table.
func NewTChatQuickReplyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatQuickReplyModel {
	return &customTChatQuickReplyModel{
		defaultTChatQuickReplyModel: newTChatQuickReplyModel(conn, c, opts...),
	}
}

func (m *customTChatQuickReplyModel) FindPage(ctx context.Context, filter ChatQuickReplyPageFilter, cursor int64, limit int64) ([]*TChatQuickReply, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(title LIKE ? OR content LIKE ?)", like, like)
	}
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("agent_id", filter.AgentId)
	builder.EqInt64("category_id", filter.CategoryId)
	builder.EqInt64("enabled", filter.Enabled)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY sort ASC,id DESC LIMIT ?,?", tChatQuickReplyRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatQuickReply
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatQuickReplyModel) ListEnabled(ctx context.Context, merchantId int64, agentId int64, categoryId int64) ([]*TChatQuickReply, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("merchant_id", merchantId)
	builder.EqInt64("agent_id", agentId)
	builder.EqInt64("category_id", categoryId)
	builder.And("enabled = ?", 1)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY sort ASC,id DESC", tChatQuickReplyRows, m.table, builder.Where())
	var list []*TChatQuickReply
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
