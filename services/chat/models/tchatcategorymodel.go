package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TChatCategoryModel = (*customTChatCategoryModel)(nil)

type (
	ChatCategoryPageFilter struct {
		Keyword      string
		MerchantId   int64
		ParentId     int64
		CategoryCode string
		CategoryName string
		Enabled      int64
	}

	// TChatCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTChatCategoryModel.
	TChatCategoryModel interface {
		tChatCategoryModel
		FindPage(ctx context.Context, filter ChatCategoryPageFilter, cursor int64, limit int64) ([]*TChatCategory, int64, error)
		ListEnabledByMerchant(ctx context.Context, merchantId int64) ([]*TChatCategory, error)
	}

	customTChatCategoryModel struct {
		*defaultTChatCategoryModel
	}
)

// NewTChatCategoryModel returns a model for the database table.
func NewTChatCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TChatCategoryModel {
	return &customTChatCategoryModel{
		defaultTChatCategoryModel: newTChatCategoryModel(conn, c, opts...),
	}
}

func (m *customTChatCategoryModel) FindPage(ctx context.Context, filter ChatCategoryPageFilter, cursor int64, limit int64) ([]*TChatCategory, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		builder.And("(category_code LIKE ? OR category_name LIKE ?)", like, like)
	}
	builder.EqInt64("merchant_id", filter.MerchantId)
	builder.EqInt64("parent_id", filter.ParentId)
	builder.EqString("category_code", filter.CategoryCode)
	builder.LikeString("category_name", filter.CategoryName)
	builder.EqInt64("enabled", filter.Enabled)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY sort ASC,id DESC LIMIT ?,?", tChatCategoryRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*TChatCategory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTChatCategoryModel) ListEnabledByMerchant(ctx context.Context, merchantId int64) ([]*TChatCategory, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE merchant_id = ? AND enabled = ? ORDER BY sort ASC,id DESC", tChatCategoryRows, m.table)
	var list []*TChatCategory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, merchantId, 1); err != nil {
		return nil, err
	}
	return list, nil
}
