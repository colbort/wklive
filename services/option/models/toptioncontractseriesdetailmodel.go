package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionContractSeriesDetailModel = (*customTOptionContractSeriesDetailModel)(nil)

type (
	// TOptionContractSeriesDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractSeriesDetailModel.
	TOptionContractSeriesDetailModel interface {
		tOptionContractSeriesDetailModel
		FindPageBySeries(ctx context.Context, tenantId, seriesId, cursor, limit int64) ([]*TOptionContractSeriesDetail, int64, error)
		FindBySeries(ctx context.Context, tenantId, seriesId int64) ([]*TOptionContractSeriesDetail, error)
		FindSeriesLaunchByContract(ctx context.Context, tenantId, contractId int64) (*TOptionContractSeries, error)
	}

	customTOptionContractSeriesDetailModel struct {
		*defaultTOptionContractSeriesDetailModel
	}
)

func (m *defaultTOptionContractSeriesDetailModel) FindBySeries(
	ctx context.Context, tenantId, seriesId int64,
) ([]*TOptionContractSeriesDetail, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND series_id=?
ORDER BY expiry_id,strike_price,option_type,id`,
		tOptionContractSeriesDetailRows, m.table)
	var items []*TOptionContractSeriesDetail
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, seriesId); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *defaultTOptionContractSeriesDetailModel) FindSeriesLaunchByContract(
	ctx context.Context, tenantId, contractId int64,
) (*TOptionContractSeries, error) {
	query := fmt.Sprintf(`SELECT s.*
FROM t_option_contract_series s
JOIN %s d ON d.tenant_id=s.tenant_id AND d.series_id=s.id
WHERE d.tenant_id=? AND d.contract_id=? LIMIT 1`,
		m.table)
	var item TOptionContractSeries
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, contractId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionContractSeriesDetailModel) FindPageBySeries(
	ctx context.Context, tenantId, seriesId, cursor, limit int64,
) ([]*TOptionContractSeriesDetail, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id=? AND series_id=?", m.table),
		tenantId, seriesId); err != nil {
		return nil, 0, err
	}
	args := []any{tenantId, seriesId}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND series_id=?", tOptionContractSeriesDetailRows, m.table)
	if cursor > 0 {
		query += " AND id < ?"
		args = append(args, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)
	var items []*TOptionContractSeriesDetail
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionContractSeriesDetailModel returns a model for the database table.
func NewTOptionContractSeriesDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractSeriesDetailModel {
	return &customTOptionContractSeriesDetailModel{
		defaultTOptionContractSeriesDetailModel: newTOptionContractSeriesDetailModel(conn, c, opts...),
	}
}
