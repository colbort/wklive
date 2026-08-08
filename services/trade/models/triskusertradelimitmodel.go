package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TRiskUserTradeLimitModel = (*customTRiskUserTradeLimitModel)(nil)

type (
	// TRiskUserTradeLimitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskUserTradeLimitModel.
	TRiskUserTradeLimitModel interface {
		tRiskUserTradeLimitModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error)
		FindControlPage(ctx context.Context, filter UserTradeControlFilter, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TRiskUserTradeLimit, error)
	}

	UserTradeControlFilter struct {
		TenantId     int64
		UserId       int64
		ProductType  int64
		ContractType int64
		SymbolId     int64
		Enabled      int64
		ControlId    int64
		ScopeType    int64
	}

	customTRiskUserTradeLimitModel struct {
		*defaultTRiskUserTradeLimitModel
	}
)

// NewTRiskUserTradeLimitModel returns a model for the database table.
func NewTRiskUserTradeLimitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskUserTradeLimitModel {
	return &customTRiskUserTradeLimitModel{
		defaultTRiskUserTradeLimitModel: newTRiskUserTradeLimitModel(conn, c, opts...),
	}
}

func (m *customTRiskUserTradeLimitModel) FindOneForUpdate(ctx context.Context, id int64) (*TRiskUserTradeLimit, error) {
	var row TRiskUserTradeLimit
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? FOR UPDATE", tRiskUserTradeLimitRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTRiskUserTradeLimitModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
			tRiskUserTradeLimitRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tRiskUserTradeLimitRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TRiskUserTradeLimit
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTRiskUserTradeLimitModel) FindControlPage(ctx context.Context, filter UserTradeControlFilter, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("user_id", filter.UserId)
	b.EqInt64("product_type", filter.ProductType)
	if filter.ContractType > 0 {
		b.EqInt64("contract_type", filter.ContractType)
	}
	b.EqInt64("enabled", filter.Enabled)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tRiskUserTradeLimitRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TRiskUserTradeLimit
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
