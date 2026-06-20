package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatGroupModel = (*customTChatGroupModel)(nil)

type (
	ChatGroupPageFilter struct {
		Keyword    string
		MerchantId int64
		GroupCode  string
		GroupName  string
		Enabled    int64
	}

	// TChatGroupModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatGroupModel.
	TChatGroupModel interface {
		tChatGroupModel
		FindPage(ctx context.Context, filter ChatGroupPageFilter, cursor int64, limit int64) ([]*TChatGroup, int64, error)
		ListEnabledByMerchant(ctx context.Context, merchantId int64) ([]*TChatGroup, error)
	}

	customTChatGroupModel struct {
		*defaultTChatGroupModel
	}
)

// NewTChatGroupModel returns a model for the database table.
func NewTChatGroupModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatGroupModel {
	return &customTChatGroupModel{
		defaultTChatGroupModel: newTChatGroupModel(conn, c, opts...),
	}
}

func (m *customTChatGroupModel) FindPage(ctx context.Context, filter ChatGroupPageFilter, cursor int64, limit int64) ([]*TChatGroup, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(group_code LIKE ? OR group_name LIKE ?)", like, like)
	}
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqString("group_code", filter.GroupCode)
	builder.LikeString("group_name", filter.GroupName)
	builder.EqInt64("enabled", filter.Enabled)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY sort ASC,id DESC LIMIT ?,?", tChatGroupRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatGroup
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatGroupModel) ListEnabledByMerchant(ctx context.Context, merchantId int64) ([]*TChatGroup, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE merchant_id = ? AND enabled = ? ORDER BY sort ASC,id DESC", tChatGroupRows, m.table)
	var list []*TChatGroup
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, merchantId, 1); err != nil {
		return nil, err
	}
	return list, nil
}
