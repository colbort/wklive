package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
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
		AddConsumed(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error)
		AddReleased(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error)
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

func (m *defaultTTradeAssetReservationModel) AddConsumed(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error) {
	return m.addSettledAmount(ctx, id, "consumed_amount", amount, updateTimes)
}

func (m *defaultTTradeAssetReservationModel) AddReleased(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error) {
	return m.addSettledAmount(ctx, id, "released_amount", amount, updateTimes)
}

func (m *defaultTTradeAssetReservationModel) addSettledAmount(ctx context.Context, id int64, column string, amount decimal.Decimal, updateTimes int64) (bool, error) {
	if id <= 0 || !amount.IsPositive() || (column != "consumed_amount" && column != "released_amount") {
		return false, nil
	}
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeAssetReservationIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeAssetReservationTenantIdReservationNoPrefix, item.TenantId, item.ReservationNo)
	statusSQL := "CASE WHEN consumed_amount + ? + released_amount = reserved_amount THEN 4 ELSE 3 END"
	if column == "released_amount" {
		statusSQL = "CASE WHEN consumed_amount + released_amount + ? = reserved_amount THEN 6 ELSE 3 END"
	}
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET %s = %s + ?, status = %s, version = version + 1, update_times = ? WHERE id = ? AND consumed_amount + released_amount + ? <= reserved_amount", m.table, column, column, statusSQL)
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, id, amount)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
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
