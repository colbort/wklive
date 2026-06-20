package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractPositionModel = (*customTContractPositionModel)(nil)

type (
	ContractPositionPageFilter struct {
		TenantId     int64
		UserId       int64
		SymbolId     int64
		MarketType   int64
		PositionSide int64
	}

	// TContractPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionModel.
	TContractPositionModel interface {
		tContractPositionModel
		FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error)
		FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error)
	}

	customTContractPositionModel struct {
		*defaultTContractPositionModel
	}
)

// NewTContractPositionModel returns a model for the database table.
func NewTContractPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionModel {
	return &customTContractPositionModel{
		defaultTContractPositionModel: newTContractPositionModel(conn, c, opts...),
	}
}

func (m *defaultTContractPositionModel) FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("position_side", filter.PositionSide)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractPositionRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTContractPositionModel) FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqInt64("position_side", filter.PositionSide)

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC", tContractPositionRows, m.table, builder.Where())
	var list []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
