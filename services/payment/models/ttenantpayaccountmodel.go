package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTenantPayAccountModel = (*customTTenantPayAccountModel)(nil)

type (
	TenantPayAccountPageFilter struct {
		TenantId   int64
		PlatformId int64
	}

	// TTenantPayAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayAccountModel.
	TTenantPayAccountModel interface {
		tTenantPayAccountModel
		FindPage(ctx context.Context, filter TenantPayAccountPageFilter, cursor int64, limit int64) ([]*TTenantPayAccount, int64, error)
	}

	customTTenantPayAccountModel struct {
		*defaultTTenantPayAccountModel
	}
)

// NewTTenantPayAccountModel returns a model for the database table.
func NewTTenantPayAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayAccountModel {
	return &customTTenantPayAccountModel{
		defaultTTenantPayAccountModel: newTTenantPayAccountModel(conn, c, opts...),
	}
}

func (m *defaultTTenantPayAccountModel) FindPage(ctx context.Context, filter TenantPayAccountPageFilter, cursor int64, limit int64) ([]*TTenantPayAccount, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("platform_id", filter.PlatformId)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayAccountRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTenantPayAccount
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
