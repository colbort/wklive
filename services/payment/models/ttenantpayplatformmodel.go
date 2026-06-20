package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTenantPayPlatformModel = (*customTTenantPayPlatformModel)(nil)

type (
	TenantPayPlatformPageFilter struct {
		TenantId   int64
		PlatformId int64
		Enabled    int64
		OpenStatus int64
	}

	// TTenantPayPlatformModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTenantPayPlatformModel.
	TTenantPayPlatformModel interface {
		tTenantPayPlatformModel
		FindPage(ctx context.Context, filter TenantPayPlatformPageFilter, cursor int64, limit int64) ([]*TTenantPayPlatform, int64, error)
	}

	customTTenantPayPlatformModel struct {
		*defaultTTenantPayPlatformModel
	}
)

// NewTTenantPayPlatformModel returns a model for the database table.
func NewTTenantPayPlatformModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTenantPayPlatformModel {
	return &customTTenantPayPlatformModel{
		defaultTTenantPayPlatformModel: newTTenantPayPlatformModel(conn, c, opts...),
	}
}

func (m *defaultTTenantPayPlatformModel) FindPage(ctx context.Context, filter TenantPayPlatformPageFilter, cursor int64, limit int64) ([]*TTenantPayPlatform, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("platform_id", filter.PlatformId)
	builder.EqInt64("enabled", filter.Enabled)
	builder.EqInt64("open_status", filter.OpenStatus)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTenantPayPlatformRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTenantPayPlatform
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
