package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TPayProductModel = (*customTPayProductModel)(nil)

type (
	// TPayProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayProductModel.
	TPayProductModel interface {
		tPayProductModel
		FindPage(ctx context.Context, platformId int64, cursor int64, limit int64) ([]*TPayProduct, int64, error)
	}

	customTPayProductModel struct {
		*defaultTPayProductModel
	}
)

// NewTPayProductModel returns a model for the database table.
func NewTPayProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TPayProductModel {
	return &customTPayProductModel{
		defaultTPayProductModel: newTPayProductModel(conn, c, opts...),
	}
}

func (m *defaultTPayProductModel) FindPage(ctx context.Context, platformId int64, cursor int64, limit int64) ([]*TPayProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("platform_id", platformId)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tPayProductRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TPayProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
