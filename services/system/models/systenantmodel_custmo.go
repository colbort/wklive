package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type TenantModel interface {
	sysTenantModel
	FindPage(ctx context.Context, keyword string, status int64, tenantName string, tenantCode string, contactName string, contactPhone string, cursor int64, limit int64) ([]*SysTenant, int64, error)
	FindByTenantCode(ctx context.Context, tenantCode string) (*SysTenant, error)
}

func (m *customSysTenantModel) FindPage(ctx context.Context, keyword string, status int64, tenantName string, tenantCode string, contactName string, contactPhone string, cursor int64, limit int64) ([]*SysTenant, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if keyword != "" {
		like := "%" + keyword + "%"
		builder.And("(tenant_name LIKE ? OR tenant_code LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?)", like, like, like, like)
	}
	builder.EqInt64("status", status)
	builder.LikeString("tenant_name", "%"+tenantName+"%")
	builder.LikeString("tenant_code", "%"+tenantCode+"%")
	builder.LikeString("contact_name", "%"+contactName+"%")
	builder.LikeString("contact_phone", "%"+contactPhone+"%")

	where := builder.Where()
	args := builder.Args()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?,?", sysTenantRows, m.table, where)
	listArgs := append(append([]any{}, args...), cursor, limit)

	var list []*SysTenant
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	err = m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customSysTenantModel) FindByTenantCode(ctx context.Context, tenantCode string) (*SysTenant, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_code = ?", tenantCode)

	query := fmt.Sprintf("select %s from %s where %s limit 1", sysTenantRows, m.table, builder.Where())
	var resp SysTenant
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, builder.Args()...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
