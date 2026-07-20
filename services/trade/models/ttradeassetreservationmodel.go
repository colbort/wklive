package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeAssetReservationModel = (*customTTradeAssetReservationModel)(nil)

type (
	// TTradeAssetReservationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeAssetReservationModel.
	TTradeAssetReservationModel interface {
		tTradeAssetReservationModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TTradeAssetReservation, int64, error)
	}

	customTTradeAssetReservationModel struct {
		*defaultTTradeAssetReservationModel
	}
)

// NewTTradeAssetReservationModel returns a model for the database table.
func NewTTradeAssetReservationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeAssetReservationModel {
	return &customTTradeAssetReservationModel{
		defaultTTradeAssetReservationModel: newTTradeAssetReservationModel(conn, c, opts...),
	}
}

func (m *defaultTTradeAssetReservationModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TTradeAssetReservation, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeAssetReservationRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TTradeAssetReservation
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
