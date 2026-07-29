package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketCategoryModel = (*customTMarketCategoryModel)(nil)

type (
	MarketCategoryPageFilter struct {
		CategoryType int32
		Enabled      int32
		AppVisible   int32
	}

	// TMarketCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketCategoryModel.
	TMarketCategoryModel interface {
		tMarketCategoryModel
		FindAll(ctx context.Context) ([]*TMarketCategory, error)
		FindPage(ctx context.Context, filter MarketCategoryPageFilter, cursor int64, limit int64) ([]*TMarketCategory, int64, error)
	}

	customTMarketCategoryModel struct {
		*defaultTMarketCategoryModel
	}
)

// NewTMarketCategoryModel returns a model for the database table.
func NewTMarketCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketCategoryModel {
	return &customTMarketCategoryModel{
		defaultTMarketCategoryModel: newTMarketCategoryModel(conn, c, opts...),
	}
}

func (m *defaultTMarketCategoryModel) FindAll(ctx context.Context) ([]*TMarketCategory, error) {
	query := fmt.Sprintf("select %s from %s", tMarketCategoryRows, m.table)
	var resp []*TMarketCategory
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query)
	return resp, err
}

func (m *defaultTMarketCategoryModel) FindPage(ctx context.Context, filter MarketCategoryPageFilter, cursor int64, limit int64) ([]*TMarketCategory, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("category_type", int64(filter.CategoryType))
	builder.EqInt64("enabled", int64(filter.Enabled))
	builder.EqInt64("app_visible", int64(filter.AppVisible))

	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			tMarketCategoryRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tMarketCategoryRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TMarketCategory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
