package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractFundingBatchModel = (*customTContractFundingBatchModel)(nil)

type (
	// TContractFundingBatchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractFundingBatchModel.
	TContractFundingBatchModel interface {
		tContractFundingBatchModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractFundingBatch, int64, error)
		FindLatestBySymbol(ctx context.Context, tenantID, symbolID int64) (*TContractFundingBatch, error)
	}

	customTContractFundingBatchModel struct {
		*defaultTContractFundingBatchModel
	}
)

// NewTContractFundingBatchModel returns a model for the database table.
func NewTContractFundingBatchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractFundingBatchModel {
	return &customTContractFundingBatchModel{
		defaultTContractFundingBatchModel: newTContractFundingBatchModel(conn, c, opts...),
	}
}

func (m *customTContractFundingBatchModel) FindLatestBySymbol(ctx context.Context, tenantID, symbolID int64) (*TContractFundingBatch, error) {
	var row TContractFundingBatch
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND symbol_id = ? ORDER BY settlement_time DESC, id DESC LIMIT 1", tContractFundingBatchRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, symbolID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTContractFundingBatchModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractFundingBatch, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "settlement_time")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractFundingBatchRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TContractFundingBatch
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
