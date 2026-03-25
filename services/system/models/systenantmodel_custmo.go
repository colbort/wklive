package models

import "context"

type TenantModel interface {
	sysTenantModel
	FindPage(ctx context.Context, keyword string, status int64, tenantName string, tenantCode string, contactName string, contactPhone string, cursor int64, limit int64) ([]*SysTenant, int64, error)
	FindByTenantCode(ctx context.Context, tenantCode string) (*SysTenant, error)
}

func (m *customSysTenantModel) FindPage(ctx context.Context, keyword string, status int64, tenantName string, tenantCode string, contactName string, contactPhone string, cursor int64, limit int64) ([]*SysTenant, int64, error) {
	query := "select " + sysTenantRows + " from " + m.table + " where 1=1"
	var args []interface{}
	if keyword != "" {
		query += " and (tenant_name like ? or tenant_code like ? or contact_name like ? or contact_phone like ?)"
		keyword = "%" + keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword)
	}
	if status != 0 {
		query += " and status = ?"
		args = append(args, status)
	}
	if tenantName != "" {
		query += " and tenant_name like ?"
		tenantName = "%" + tenantName + "%"
		args = append(args, tenantName)
	}
	if tenantCode != "" {
		query += " and tenant_code like ?"
		tenantCode = "%" + tenantCode + "%"
		args = append(args, tenantCode)
	}
	if contactName != "" {
		query += " and contact_name like ?"
		contactName = "%" + contactName + "%"
		args = append(args, contactName)
	}
	if contactPhone != "" {
		query += " and contact_phone like ?"
		contactPhone = "%" + contactPhone + "%"
		args = append(args, contactPhone)
	}
	query += " order by id desc limit ?,?"
	args = append(args, cursor, limit)

	var list []*SysTenant
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	countQuery := "select count(1) from " + m.table + " where 1=1"
	if keyword != "" {
		countQuery += " and (tenant_name like ? or tenant_code like ? or contact_name like ? or contact_phone like ?)"
	}
	if status != 0 {
		countQuery += " and status = ?"
	}
	if tenantName != "" {
		countQuery += " and tenant_name like ?"
	}
	if tenantCode != "" {
		countQuery += " and tenant_code like ?"
	}
	if contactName != "" {
		countQuery += " and contact_name like ?"
	}
	if contactPhone != "" {
		countQuery += " and contact_phone like ?"
	}

	err = m.QueryRowNoCacheCtx(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customSysTenantModel) FindByTenantCode(ctx context.Context, tenantCode string) (*SysTenant, error) {
	query := "select " + sysTenantRows + " from " + m.table + " where tenant_code = ? limit 1"
	var resp SysTenant
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantCode)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
