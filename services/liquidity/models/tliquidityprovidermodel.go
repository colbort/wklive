package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityProviderModel = (*customTLiquidityProviderModel)(nil)

type (
	LiquidityProviderPageFilter struct {
		ProviderType int64
		Status       int64
		Keyword      string
	}
	// TLiquidityProviderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityProviderModel.
	TLiquidityProviderModel interface {
		tLiquidityProviderModel
		FindPage(ctx context.Context, filter LiquidityProviderPageFilter, cursor, limit int64) ([]*TLiquidityProvider, int64, error)
		FindOneByProviderTypeTradeUserId(ctx context.Context, providerType, tradeUserID int64) (*TLiquidityProvider, error)
	}

	customTLiquidityProviderModel struct {
		*defaultTLiquidityProviderModel
	}
)

// NewTLiquidityProviderModel returns a model for the database table.
func NewTLiquidityProviderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityProviderModel {
	return &customTLiquidityProviderModel{
		defaultTLiquidityProviderModel: newTLiquidityProviderModel(conn, c, opts...),
	}
}

func (m *customTLiquidityProviderModel) FindOneByProviderTypeTradeUserId(ctx context.Context, providerType, tradeUserID int64) (*TLiquidityProvider, error) {
	var row TLiquidityProvider
	query := fmt.Sprintf("SELECT %s FROM %s WHERE provider_type = ? AND trade_user_id = ? LIMIT 1", tLiquidityProviderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &row, query, providerType, tradeUserID)
	if err == sqlc.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquidityProviderModel) FindPage(ctx context.Context, filter LiquidityProviderPageFilter, cursor, limit int64) ([]*TLiquidityProvider, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("provider_type", filter.ProviderType)
	b.EqInt64("status", filter.Status)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		b.And("(provider_code LIKE ? OR provider_name LIKE ? OR venue_code LIKE ?)", kw, kw, kw)
	}
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityProviderRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityProvider
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
