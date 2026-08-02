package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionContractSeriesModel = (*customTOptionContractSeriesModel)(nil)

type (
	OptionContractSeriesPageFilter struct {
		TenantId   int64
		SeriesCode string
		Status     int64
	}

	// TOptionContractSeriesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionContractSeriesModel.
	TOptionContractSeriesModel interface {
		tOptionContractSeriesModel
		FindOneByTenantIdRequestKeyNoCache(ctx context.Context, tenantId int64, requestKey string) (*TOptionContractSeries, error)
		FindLatestForUpdate(ctx context.Context, tenantId int64, seriesCode string) (*TOptionContractSeries, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionContractSeries, error)
		FindPage(ctx context.Context, filter OptionContractSeriesPageFilter, cursor, limit int64) ([]*TOptionContractSeries, int64, error)
	}

	customTOptionContractSeriesModel struct {
		*defaultTOptionContractSeriesModel
	}
)

func (m *defaultTOptionContractSeriesModel) FindOneByTenantIdRequestKeyNoCache(
	ctx context.Context, tenantId int64, requestKey string,
) (*TOptionContractSeries, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND request_key=? LIMIT 1",
		tOptionContractSeriesRows, m.table)
	var item TOptionContractSeries
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, requestKey); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionContractSeriesModel) FindLatestForUpdate(
	ctx context.Context, tenantId int64, seriesCode string,
) (*TOptionContractSeries, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND series_code=? ORDER BY version DESC LIMIT 1 FOR UPDATE",
		tOptionContractSeriesRows, m.table)
	var item TOptionContractSeries
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, seriesCode); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionContractSeriesModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionContractSeries, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE", tOptionContractSeriesRows, m.table)
	var item TOptionContractSeries
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionContractSeriesModel) FindPage(
	ctx context.Context, filter OptionContractSeriesPageFilter, cursor, limit int64,
) ([]*TOptionContractSeries, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("series_code", filter.SeriesCode)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionContractSeriesRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionContractSeries
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionContractSeriesModel returns a model for the database table.
func NewTOptionContractSeriesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionContractSeriesModel {
	return &customTOptionContractSeriesModel{
		defaultTOptionContractSeriesModel: newTOptionContractSeriesModel(conn, c, opts...),
	}
}
