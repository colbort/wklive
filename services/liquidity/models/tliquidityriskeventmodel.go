package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityRiskEventModel = (*customTLiquidityRiskEventModel)(nil)

type (
	LiquidityRiskEventPageFilter struct {
		TenantId, ConfigId, ProviderId int64
		RiskType                       string
		RiskLevel, Status              int64
		TimeStart, TimeEnd             int64
	}
	// TLiquidityRiskEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityRiskEventModel.
	TLiquidityRiskEventModel interface {
		tLiquidityRiskEventModel
		FindPage(ctx context.Context, filter LiquidityRiskEventPageFilter, cursor, limit int64) ([]*TLiquidityRiskEvent, int64, error)
	}

	customTLiquidityRiskEventModel struct {
		*defaultTLiquidityRiskEventModel
	}
)

// NewTLiquidityRiskEventModel returns a model for the database table.
func NewTLiquidityRiskEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityRiskEventModel {
	return &customTLiquidityRiskEventModel{
		defaultTLiquidityRiskEventModel: newTLiquidityRiskEventModel(conn, c, opts...),
	}
}

func (m *customTLiquidityRiskEventModel) FindPage(ctx context.Context, filter LiquidityRiskEventPageFilter, cursor, limit int64) ([]*TLiquidityRiskEvent, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqString("risk_type", filter.RiskType)
	b.EqInt64("risk_level", filter.RiskLevel)
	b.EqInt64("status", filter.Status)
	b.GteInt64("triggered_at", filter.TimeStart)
	b.LteInt64("triggered_at", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityRiskEventRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityRiskEvent
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
