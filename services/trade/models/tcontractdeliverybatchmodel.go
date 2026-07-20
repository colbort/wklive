package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractDeliveryBatchModel = (*customTContractDeliveryBatchModel)(nil)

type (
	// TContractDeliveryBatchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractDeliveryBatchModel.
	TContractDeliveryBatchModel interface {
		tContractDeliveryBatchModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractDeliveryBatch, int64, error)
	}

	customTContractDeliveryBatchModel struct {
		*defaultTContractDeliveryBatchModel
	}
)

// NewTContractDeliveryBatchModel returns a model for the database table.
func NewTContractDeliveryBatchModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractDeliveryBatchModel {
	return &customTContractDeliveryBatchModel{
		defaultTContractDeliveryBatchModel: newTContractDeliveryBatchModel(conn, c, opts...),
	}
}

func (m *defaultTContractDeliveryBatchModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractDeliveryBatch, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "delivery_time")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractDeliveryBatchRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TContractDeliveryBatch
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
