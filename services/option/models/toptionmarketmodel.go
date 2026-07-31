package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarketModel = (*customTOptionMarketModel)(nil)

type (
	// TOptionMarketModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarketModel.
	TOptionMarketModel interface {
		tOptionMarketModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TOptionMarket, int64, error)
		FindOneByTenantIdContractIdForUpdate(ctx context.Context, tenantId, contractId int64) (*TOptionMarket, error)
		FindByContractIDs(ctx context.Context, tenantId int64, contractIDs []int64) ([]*TOptionMarket, error)
	}

	customTOptionMarketModel struct {
		*defaultTOptionMarketModel
	}
)

func (m *defaultTOptionMarketModel) FindByContractIDs(
	ctx context.Context, tenantId int64, contractIDs []int64,
) ([]*TOptionMarket, error) {
	if len(contractIDs) == 0 {
		return []*TOptionMarket{}, nil
	}
	args := make([]any, 0, len(contractIDs)+1)
	args = append(args, tenantId)
	for _, id := range contractIDs {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(contractIDs)), ",")
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id=? AND contract_id IN (%s)",
		tOptionMarketRows, m.table, placeholders,
	)
	var items []*TOptionMarket
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...)
	return items, err
}

func (m *defaultTOptionMarketModel) FindOneByTenantIdContractIdForUpdate(
	ctx context.Context, tenantId, contractId int64,
) (*TOptionMarket, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND contract_id = ? LIMIT 1 FOR UPDATE`, tOptionMarketRows, m.table)
	var item TOptionMarket
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, contractId); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionMarketModel returns a model for the database table.
func NewTOptionMarketModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarketModel {
	return &customTOptionMarketModel{
		defaultTOptionMarketModel: newTOptionMarketModel(conn, c, opts...),
	}
}

func (m *defaultTOptionMarketModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TOptionMarket, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 2)

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
			tOptionMarketRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tOptionMarketRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TOptionMarket
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
