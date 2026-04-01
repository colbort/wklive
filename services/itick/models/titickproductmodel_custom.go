package models

import (
	"context"
	"fmt"
)

type ItickProductModel interface {
	tItickProductModel
	FindPage(ctx context.Context, categoryType int64, market string, appVisible int64, enabled int64, cursor int64, limit int64) ([]*TItickProduct, int64, error)
}

func (m *defaultTItickProductModel) FindPage(ctx context.Context, categoryType int64, market string, appVisible int64, enabled int64, cursor int64, limit int64) ([]*TItickProduct, int64, error) {
	query := fmt.Sprintf("select %s from %s where category_type = ? and market = ? and app_visible = ? and enabled = ? and id > ? order by id limit ?", tItickProductRows, m.table)
	var resp []*TItickProduct
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, categoryType, market, appVisible, enabled, cursor, limit)
	if err != nil {
		return nil, 0, err
	}
	var nextCursor int64
	if len(resp) > 0 {
		nextCursor = resp[len(resp)-1].Id
	}
	return resp, nextCursor, nil
}
