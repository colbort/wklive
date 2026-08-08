package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TPayProductModel = (*customTPayProductModel)(nil)

type (
	// TPayProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTPayProductModel.
	TPayProductModel interface {
		tPayProductModel
		FindPage(ctx context.Context, platformId int64, cursor int64, limit int64) ([]*TPayProduct, int64, error)
		FindByIDs(ctx context.Context, ids []int64) ([]*TPayProduct, error)
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

func (m *customTPayProductModel) FindByIDs(ctx context.Context, ids []int64) ([]*TPayProduct, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	builder := sqlutil.NewPageQueryBuilder()
	builder.InInt64("id", ids)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tPayProductRows, m.table, builder.Where())
	var list []*TPayProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTPayProductModel) FindPage(ctx context.Context, platformId int64, cursor int64, limit int64) ([]*TPayProduct, int64, error) {
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
