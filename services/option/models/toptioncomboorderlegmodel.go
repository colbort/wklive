package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionComboOrderLegModel = (*customTOptionComboOrderLegModel)(nil)

type (
	// TOptionComboOrderLegModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionComboOrderLegModel.
	TOptionComboOrderLegModel interface {
		tOptionComboOrderLegModel
		FindByComboOrderID(ctx context.Context, tenantId, comboOrderId int64) ([]*TOptionComboOrderLeg, error)
		FindByComboOrderIDForUpdate(ctx context.Context, tenantId, comboOrderId int64) ([]*TOptionComboOrderLeg, error)
	}

	customTOptionComboOrderLegModel struct {
		*defaultTOptionComboOrderLegModel
	}
)

// NewTOptionComboOrderLegModel returns a model for the database table.
func NewTOptionComboOrderLegModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionComboOrderLegModel {
	return &customTOptionComboOrderLegModel{
		defaultTOptionComboOrderLegModel: newTOptionComboOrderLegModel(conn, c, opts...),
	}
}

func (m *defaultTOptionComboOrderLegModel) FindByComboOrderID(
	ctx context.Context, tenantId, comboOrderId int64,
) ([]*TOptionComboOrderLeg, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND combo_order_id=?
ORDER BY leg_no`, tOptionComboOrderLegRows, m.table)
	var list []*TOptionComboOrderLeg
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTOptionComboOrderLegModel) FindByComboOrderIDForUpdate(
	ctx context.Context, tenantId, comboOrderId int64,
) ([]*TOptionComboOrderLeg, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND combo_order_id=?
ORDER BY leg_no FOR UPDATE`, tOptionComboOrderLegRows, m.table)
	var list []*TOptionComboOrderLeg
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId); err != nil {
		return nil, err
	}
	return list, nil
}
