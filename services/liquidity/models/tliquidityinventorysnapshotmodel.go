package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityInventorySnapshotModel = (*customTLiquidityInventorySnapshotModel)(nil)

type (
	LiquidityInventorySnapshotPageFilter struct {
		TenantId, ConfigId, ProviderId, Source int64
		TimeStart, TimeEnd                     int64
	}
	// TLiquidityInventorySnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityInventorySnapshotModel.
	TLiquidityInventorySnapshotModel interface {
		tLiquidityInventorySnapshotModel
		FindPage(ctx context.Context, filter LiquidityInventorySnapshotPageFilter, cursor, limit int64) ([]*TLiquidityInventorySnapshot, int64, error)
		FindLatest(ctx context.Context, tenantID, configID, providerID, source int64) (*TLiquidityInventorySnapshot, error)
	}

	customTLiquidityInventorySnapshotModel struct {
		*defaultTLiquidityInventorySnapshotModel
	}
)

// NewTLiquidityInventorySnapshotModel returns a model for the database table.
func NewTLiquidityInventorySnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityInventorySnapshotModel {
	return &customTLiquidityInventorySnapshotModel{
		defaultTLiquidityInventorySnapshotModel: newTLiquidityInventorySnapshotModel(conn, c, opts...),
	}
}

func (m *customTLiquidityInventorySnapshotModel) FindPage(ctx context.Context, filter LiquidityInventorySnapshotPageFilter, cursor, limit int64) ([]*TLiquidityInventorySnapshot, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("source", filter.Source)
	b.GteInt64("snapshot_time", filter.TimeStart)
	b.LteInt64("snapshot_time", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityInventorySnapshotRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityInventorySnapshot
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *customTLiquidityInventorySnapshotModel) FindLatest(ctx context.Context, tenantID, configID, providerID, source int64) (*TLiquidityInventorySnapshot, error) {
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", tenantID)
	b.EqInt64("config_id", configID)
	b.EqInt64("provider_id", providerID)
	b.EqInt64("source", source)
	var row TLiquidityInventorySnapshot
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY snapshot_time DESC, id DESC LIMIT 1", tLiquidityInventorySnapshotRows, m.table, b.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, b.Args()...); err != nil {
		return nil, err
	}
	return &row, nil
}
