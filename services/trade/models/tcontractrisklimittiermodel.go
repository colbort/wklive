package models

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractRiskLimitTierModel = (*customTContractRiskLimitTierModel)(nil)

type (
	// TContractRiskLimitTierModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractRiskLimitTierModel.
	TContractRiskLimitTierModel interface {
		tContractRiskLimitTierModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractRiskLimitTier, int64, error)
		FindByNotional(ctx context.Context, tenantId, symbolId int64, notional decimal.Decimal) (*TContractRiskLimitTier, error)
		CountOverlapping(ctx context.Context, tenantID, symbolID, excludeID, excludeTierNo int64, floor, capValue decimal.Decimal) (int64, error)
	}

	customTContractRiskLimitTierModel struct {
		*defaultTContractRiskLimitTierModel
	}
)

func (m *defaultTContractRiskLimitTierModel) CountOverlapping(
	ctx context.Context,
	tenantID, symbolID, excludeID, excludeTierNo int64,
	floor, capValue decimal.Decimal,
) (int64, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_contract_risk_limit_tier
WHERE tenant_id=? AND symbol_id=? AND id<>? AND tier_no<>?
  AND (notional_cap=0 OR notional_cap>?)
  AND (?=0 OR notional_floor<?)`,
		tenantID, symbolID, excludeID, excludeTierNo, floor, capValue, capValue)
	return count, err
}

// NewTContractRiskLimitTierModel returns a model for the database table.
func NewTContractRiskLimitTierModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractRiskLimitTierModel {
	return &customTContractRiskLimitTierModel{
		defaultTContractRiskLimitTierModel: newTContractRiskLimitTierModel(conn, c, opts...),
	}
}

func (m *defaultTContractRiskLimitTierModel) FindByNotional(ctx context.Context, tenantId, symbolId int64, notional decimal.Decimal) (*TContractRiskLimitTier, error) {
	var item TContractRiskLimitTier
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND symbol_id = ? AND enabled = 1 AND notional_floor <= ? AND (notional_cap = 0 OR notional_cap >= ?) ORDER BY tier_no ASC LIMIT 1", tContractRiskLimitTierRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, symbolId, notional, notional); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (m *defaultTContractRiskLimitTierModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractRiskLimitTier, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractRiskLimitTierRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TContractRiskLimitTier
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
