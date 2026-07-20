package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractDeliverySettlementModel = (*customTContractDeliverySettlementModel)(nil)

type (
	// TContractDeliverySettlementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractDeliverySettlementModel.
	TContractDeliverySettlementModel interface {
		tContractDeliverySettlementModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractDeliverySettlement, int64, error)
	}

	customTContractDeliverySettlementModel struct {
		*defaultTContractDeliverySettlementModel
	}
)

// NewTContractDeliverySettlementModel returns a model for the database table.
func NewTContractDeliverySettlementModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractDeliverySettlementModel {
	return &customTContractDeliverySettlementModel{
		defaultTContractDeliverySettlementModel: newTContractDeliverySettlementModel(conn, c, opts...),
	}
}

func (m *defaultTContractDeliverySettlementModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractDeliverySettlement, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractDeliverySettlementRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TContractDeliverySettlement
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
