package models

import (
	"context"
	"fmt"
)

type ItickTenantProductModel interface {
	tItickTenantProductModel
	FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TItickTenantProduct, int64, error)
}

func (m *defaultTItickTenantProductModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TItickTenantProduct, int64, error) {
	query := fmt.Sprintf("select %s from %s where tenant_id = ? and id > ? order by id limit ?", tItickTenantProductRows, m.table)
	var resp []*TItickTenantProduct
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, tenantId, cursor, limit)
	if err != nil {
		return nil, 0, err
	}
	var nextCursor int64
	if len(resp) > 0 {
		nextCursor = resp[len(resp)-1].Id
	}
	return resp, nextCursor, nil
}
