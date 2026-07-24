package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityReconcileDetailModel = (*customTLiquidityReconcileDetailModel)(nil)

type (
	LiquidityReconcileDetailPageFilter struct {
		TenantId, BatchId, DifferenceType, Status int64
	}
	// TLiquidityReconcileDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityReconcileDetailModel.
	TLiquidityReconcileDetailModel interface {
		tLiquidityReconcileDetailModel
		FindPage(ctx context.Context, filter LiquidityReconcileDetailPageFilter, cursor, limit int64) ([]*TLiquidityReconcileDetail, int64, error)
	}

	customTLiquidityReconcileDetailModel struct {
		*defaultTLiquidityReconcileDetailModel
	}
)

// NewTLiquidityReconcileDetailModel returns a model for the database table.
func NewTLiquidityReconcileDetailModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityReconcileDetailModel {
	return &customTLiquidityReconcileDetailModel{
		defaultTLiquidityReconcileDetailModel: newTLiquidityReconcileDetailModel(conn, c, opts...),
	}
}

func (m *customTLiquidityReconcileDetailModel) FindPage(ctx context.Context, filter LiquidityReconcileDetailPageFilter, cursor, limit int64) ([]*TLiquidityReconcileDetail, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("batch_id", filter.BatchId)
	b.EqInt64("difference_type", filter.DifferenceType)
	b.EqInt64("status", filter.Status)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityReconcileDetailRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityReconcileDetail
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
