package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementDetailModel = (*customTOptionSettlementDetailModel)(nil)

type (
	// TOptionSettlementDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementDetailModel.
	TOptionSettlementDetailModel interface {
		tOptionSettlementDetailModel
		FindByBatch(ctx context.Context, tenantId, batchId int64) ([]*TOptionSettlementDetail, error)
	}

	customTOptionSettlementDetailModel struct {
		*defaultTOptionSettlementDetailModel
	}
)

func (m *defaultTOptionSettlementDetailModel) FindByBatch(
	ctx context.Context,
	tenantId, batchId int64,
) ([]*TOptionSettlementDetail, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND batch_id = ? ORDER BY id`, tOptionSettlementDetailRows, m.table)
	var items []*TOptionSettlementDetail
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, batchId); err != nil {
		return nil, err
	}
	return items, nil
}

// NewTOptionSettlementDetailModel returns a model for the database table.
func NewTOptionSettlementDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementDetailModel {
	return &customTOptionSettlementDetailModel{
		defaultTOptionSettlementDetailModel: newTOptionSettlementDetailModel(conn, c, opts...),
	}
}
