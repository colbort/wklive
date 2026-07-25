package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/liquidity"

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
		FindRunningInternal(ctx context.Context, configID int64, limit int64) ([]*TLiquiditySymbolConfig, error)
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

func (m *customTLiquiditySymbolConfigModel) FindRunningInternal(ctx context.Context, configID int64, limit int64) ([]*TLiquiditySymbolConfig, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = ? AND liquidity_mode IN (?, ?)", tLiquiditySymbolConfigRows, m.table)
	args := []any{
		int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING),
		int64(liquidity.LiquidityMode_LIQUIDITY_MODE_INTERNAL_MARKET_MAKING),
		int64(liquidity.LiquidityMode_LIQUIDITY_MODE_INTERNAL_WITH_EXTERNAL_HEDGE),
	}
	if configID > 0 {
		query += " AND id = ?"
		args = append(args, configID)
	}
	query += " ORDER BY id ASC LIMIT ?"
	args = append(args, limit)
	var rows []*TLiquiditySymbolConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
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
	query := fmt.Sprintf("SELECT %s FROM %s WHERE symbol_id = ? AND status = ? LIMIT 1", tLiquiditySymbolConfigRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, symbolID, int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING)); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquiditySymbolConfigModel) FindActiveExternalBySymbol(ctx context.Context, symbolID int64) (*TLiquiditySymbolConfig, error) {
	var row TLiquiditySymbolConfig
	query := fmt.Sprintf("SELECT %s FROM %s WHERE symbol_id = ? AND status = ? AND liquidity_mode IN (?, ?) LIMIT 1", tLiquiditySymbolConfigRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query,
		symbolID,
		int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING),
		int64(liquidity.LiquidityMode_LIQUIDITY_MODE_EXTERNAL_ROUTING),
		int64(liquidity.LiquidityMode_LIQUIDITY_MODE_INTERNAL_WITH_EXTERNAL_HEDGE),
	); err != nil {
		return nil, err
	}
	return &row, nil
}
