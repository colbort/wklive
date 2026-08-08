package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionContractSeriesStrikeBandModel = (*customTOptionContractSeriesStrikeBandModel)(nil)

type (
	// TOptionContractSeriesStrikeBandModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractSeriesStrikeBandModel.
	TOptionContractSeriesStrikeBandModel interface {
		tOptionContractSeriesStrikeBandModel
		FindBySeries(ctx context.Context, tenantId, seriesId int64) ([]*TOptionContractSeriesStrikeBand, error)
	}

	customTOptionContractSeriesStrikeBandModel struct {
		*defaultTOptionContractSeriesStrikeBandModel
	}
)

func (m *customTOptionContractSeriesStrikeBandModel) FindBySeries(
	ctx context.Context, tenantId, seriesId int64,
) ([]*TOptionContractSeriesStrikeBand, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND series_id=? ORDER BY sequence_no,id",
		tOptionContractSeriesStrikeBandRows, m.table)
	var items []*TOptionContractSeriesStrikeBand
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, seriesId); err != nil {
		return nil, err
	}
	return items, nil
}

// NewTOptionContractSeriesStrikeBandModel returns a model for the database table.
func NewTOptionContractSeriesStrikeBandModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractSeriesStrikeBandModel {
	return &customTOptionContractSeriesStrikeBandModel{
		defaultTOptionContractSeriesStrikeBandModel: newTOptionContractSeriesStrikeBandModel(conn, c, opts...),
	}
}
