package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquiditySymbolConfigModel = (*customTLiquiditySymbolConfigModel)(nil)

type (
	LiquiditySymbolConfigPageFilter struct {
		SymbolId, ProductType, ContractType int64
		LiquidityMode, Status               int64
		Keyword                             string
	}
	// TLiquiditySymbolConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquiditySymbolConfigModel.
	TLiquiditySymbolConfigModel interface {
		tLiquiditySymbolConfigModel
		FindPage(ctx context.Context, filter LiquiditySymbolConfigPageFilter, cursor, limit int64) ([]*TLiquiditySymbolConfig, int64, error)
		FindByIDOrSymbol(ctx context.Context, id, symbolID int64) (*TLiquiditySymbolConfig, error)
		FindActiveBySymbol(ctx context.Context, symbolID int64) (*TLiquiditySymbolConfig, error)
		FindActiveExternalBySymbol(ctx context.Context, symbolID int64) (*TLiquiditySymbolConfig, error)
	}

	customTLiquiditySymbolConfigModel struct {
		*defaultTLiquiditySymbolConfigModel
	}
)

// NewTLiquiditySymbolConfigModel returns a model for the database table.
func NewTLiquiditySymbolConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquiditySymbolConfigModel {
	return &customTLiquiditySymbolConfigModel{
		defaultTLiquiditySymbolConfigModel: newTLiquiditySymbolConfigModel(conn, c, opts...),
	}
}

func (m *customTLiquiditySymbolConfigModel) FindPage(ctx context.Context, filter LiquiditySymbolConfigPageFilter, cursor, limit int64) ([]*TLiquiditySymbolConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("product_type", filter.ProductType)
	b.EqInt64("contract_type", filter.ContractType)
	b.EqInt64("liquidity_mode", filter.LiquidityMode)
	b.EqInt64("status", filter.Status)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		b.And("(symbol LIKE ? OR external_symbol LIKE ?)", kw, kw)
	}
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquiditySymbolConfigRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquiditySymbolConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *customTLiquiditySymbolConfigModel) FindByIDOrSymbol(ctx context.Context, id, symbolID int64) (*TLiquiditySymbolConfig, error) {
	var row TLiquiditySymbolConfig
	query, arg := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1", tLiquiditySymbolConfigRows, m.table), id
	if id <= 0 {
		query, arg = fmt.Sprintf("SELECT %s FROM %s WHERE symbol_id = ? LIMIT 1", tLiquiditySymbolConfigRows, m.table), symbolID
	}
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, arg); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquiditySymbolConfigModel) FindActiveBySymbol(ctx context.Context, symbolID int64) (*TLiquiditySymbolConfig, error) {
	var row TLiquiditySymbolConfig
	query := fmt.Sprintf("SELECT %s FROM %s WHERE symbol_id = ? AND status = 1 LIMIT 1", tLiquiditySymbolConfigRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, symbolID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquiditySymbolConfigModel) FindActiveExternalBySymbol(ctx context.Context, symbolID int64) (*TLiquiditySymbolConfig, error) {
	var row TLiquiditySymbolConfig
	query := fmt.Sprintf("SELECT %s FROM %s WHERE symbol_id = ? AND status = 1 AND liquidity_mode IN (2, 3) LIMIT 1", tLiquiditySymbolConfigRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, symbolID); err != nil {
		return nil, err
	}
	return &row, nil
}
