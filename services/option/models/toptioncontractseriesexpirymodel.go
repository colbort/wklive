package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionContractSeriesExpiryModel = (*customTOptionContractSeriesExpiryModel)(nil)

type (
	// TOptionContractSeriesExpiryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractSeriesExpiryModel.
	TOptionContractSeriesExpiryModel interface {
		tOptionContractSeriesExpiryModel
		FindBySeries(ctx context.Context, tenantId, seriesId int64) ([]*TOptionContractSeriesExpiry, error)
	}

	customTOptionContractSeriesExpiryModel struct {
		*defaultTOptionContractSeriesExpiryModel
	}
)

func (m *customTOptionContractSeriesExpiryModel) FindBySeries(
	ctx context.Context, tenantId, seriesId int64,
) ([]*TOptionContractSeriesExpiry, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND series_id=? ORDER BY sequence_no,id",
		tOptionContractSeriesExpiryRows, m.table)
	var items []*TOptionContractSeriesExpiry
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, seriesId); err != nil {
		return nil, err
	}
	return items, nil
}

// NewTOptionContractSeriesExpiryModel returns a model for the database table.
func NewTOptionContractSeriesExpiryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractSeriesExpiryModel {
	return &customTOptionContractSeriesExpiryModel{
		defaultTOptionContractSeriesExpiryModel: newTOptionContractSeriesExpiryModel(conn, c, opts...),
	}
}
