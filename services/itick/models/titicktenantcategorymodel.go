package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TItickTenantCategoryModel = (*customTItickTenantCategoryModel)(nil)

type (
	// TItickTenantCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickTenantCategoryModel.
	TItickTenantCategoryModel interface {
		tItickTenantCategoryModel
		FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TItickTenantCategory, int64, error)
	}

	customTItickTenantCategoryModel struct {
		*defaultTItickTenantCategoryModel
	}
)

// NewTItickTenantCategoryModel returns a model for the database table.
func NewTItickTenantCategoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickTenantCategoryModel {
	return &customTItickTenantCategoryModel{
		defaultTItickTenantCategoryModel: newTItickTenantCategoryModel(conn, c, opts...),
	}
}

func (m *defaultTItickTenantCategoryModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TItickTenantCategory, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)

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
			tItickTenantCategoryRows, m.table, where,
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
			tItickTenantCategoryRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TItickTenantCategory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
