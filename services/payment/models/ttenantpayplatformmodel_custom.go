package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type TenantPayPlatformPageFilter struct {
	TenantId   int64
	PlatformId int64
	Enabled    int64
	OpenStatus int64
}

type TenantPayPlatformModel interface {
	tTenantPayPlatformModel
	FindPage(ctx context.Context, filter TenantPayPlatformPageFilter, cursor int64, limit int64) ([]*TTenantPayPlatform, int64, error)
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
