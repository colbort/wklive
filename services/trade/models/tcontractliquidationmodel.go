package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractLiquidationModel = (*customTContractLiquidationModel)(nil)

type (
	// TContractLiquidationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractLiquidationModel.
	TContractLiquidationModel interface {
		tContractLiquidationModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractLiquidation, int64, error)
		FindActiveByPosition(ctx context.Context, tenantID, positionID int64) (*TContractLiquidation, error)
	}

	customTContractLiquidationModel struct {
		*defaultTContractLiquidationModel
	}
)

// NewTContractLiquidationModel returns a model for the database table.
func NewTContractLiquidationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractLiquidationModel {
	return &customTContractLiquidationModel{
		defaultTContractLiquidationModel: newTContractLiquidationModel(conn, c, opts...),
	}
}

func (m *defaultTContractLiquidationModel) FindActiveByPosition(ctx context.Context, tenantID, positionID int64) (*TContractLiquidation, error) {
	var row TContractLiquidation
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND position_id = ? AND status IN (?, ?, ?) ORDER BY id DESC LIMIT 1", tContractLiquidationRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, positionID, 1, 2, 4)
	switch err {
	case nil:
		return &row, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultTContractLiquidationModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractLiquidation, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "create_times")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractLiquidationRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TContractLiquidation
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
