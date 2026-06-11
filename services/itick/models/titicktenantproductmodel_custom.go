package models

import (
	"context"
	"fmt"
	"strings"
	"wklive/common/sqlutil"
)

type TenantProductPageFilter struct {
	TenantId     int64
	CategoryType int64
	Enabled      int64
	AppVisible   int64
}

type ItickTenantProductModel interface {
	tItickTenantProductModel
	FindPage(ctx context.Context, filter TenantProductPageFilter, cursor int64, limit int64) ([]*TItickTenantProduct, int64, error)
}

func (m *defaultTItickTenantProductModel) FindPage(ctx context.Context, filter TenantProductPageFilter, cursor int64, limit int64) ([]*TItickTenantProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tp.tenant_id", filter.TenantId)
	builder.EqInt64("p.category_type", filter.CategoryType)
	builder.EqInt64("tp.enabled", filter.Enabled)
	builder.EqInt64("tp.app_visible", filter.AppVisible)

	where := builder.Where()
	args := builder.Args()
	fromSql := fmt.Sprintf("%s AS tp JOIN `t_itick_product` AS p ON p.id = tp.product_id", m.table)

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", fromSql, where)
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
			ORDER BY tp.id DESC
			LIMIT ?`,
			qualifyRows("tp", tItickTenantProductRows), fromSql, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND tp.id < ?
			ORDER BY tp.id DESC
			LIMIT ?`,
			qualifyRows("tp", tItickTenantProductRows), fromSql, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TItickTenantProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func qualifyRows(alias string, rows string) string {
	fields := strings.Split(rows, ",")
	for i, field := range fields {
		fields[i] = alias + "." + strings.TrimSpace(field)
	}
	return strings.Join(fields, ",")
}
