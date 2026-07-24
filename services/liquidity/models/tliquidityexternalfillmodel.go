package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityExternalFillModel = (*customTLiquidityExternalFillModel)(nil)

type (
	LiquidityExternalFillPageFilter struct {
		TenantId, ProviderId, ExternalOrderId, SettlementStatus int64
		TimeStart, TimeEnd                                      int64
	}
	// TLiquidityExternalFillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityExternalFillModel.
	TLiquidityExternalFillModel interface {
		tLiquidityExternalFillModel
		FindPage(ctx context.Context, filter LiquidityExternalFillPageFilter, cursor, limit int64) ([]*TLiquidityExternalFill, int64, error)
	}

	customTLiquidityExternalFillModel struct {
		*defaultTLiquidityExternalFillModel
	}
)

// NewTLiquidityExternalFillModel returns a model for the database table.
func NewTLiquidityExternalFillModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityExternalFillModel {
	return &customTLiquidityExternalFillModel{
		defaultTLiquidityExternalFillModel: newTLiquidityExternalFillModel(conn, c, opts...),
	}
}

func (m *customTLiquidityExternalFillModel) FindPage(ctx context.Context, filter LiquidityExternalFillPageFilter, cursor, limit int64) ([]*TLiquidityExternalFill, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("external_order_id", filter.ExternalOrderId)
	b.EqInt64("settlement_status", filter.SettlementStatus)
	b.GteInt64("trade_time", filter.TimeStart)
	b.LteInt64("trade_time", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityExternalFillRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityExternalFill
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
