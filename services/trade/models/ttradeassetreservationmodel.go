package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
		BeginRelease(ctx context.Context, id int64, updateTimes int64) (bool, error)
		MarkSettlementFailure(ctx context.Context, id, retryStatus int64, terminal bool, nextRetryAt int64, message string, updateTimes int64) error
		FindOneByReservationNoForUpdate(ctx context.Context, tenantID int64, reservationNo string) (*TTradeAssetReservation, error)
		CountUnsettledCrossMarginByRiskUnit(ctx context.Context, tenantID, userID int64, marginAsset string) (int64, error)
	}

	customTTradeAssetReservationModel struct {
		*defaultTTradeAssetReservationModel
	}
)

func (m *customTTradeAssetReservationModel) CountUnsettledCrossMarginByRiskUnit(
	ctx context.Context, tenantID, userID int64, marginAsset string,
) (int64, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_trade_asset_reservation r
JOIN t_trade_order_contract c
  ON c.tenant_id=r.tenant_id AND c.order_id=r.order_id
JOIN t_trade_order o
  ON o.tenant_id=r.tenant_id AND o.id=r.order_id
WHERE r.tenant_id=? AND o.user_id=? AND c.margin_mode=1 AND c.margin_asset=?
  AND r.reserved_amount>r.consumed_amount+r.released_amount`,
		tenantID, userID, marginAsset)
	return count, err
}

// NewTTradeAssetReservationModel returns a model for the database table.
func NewTTradeAssetReservationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeAssetReservationModel {
	return &customTTradeAssetReservationModel{
		defaultTTradeAssetReservationModel: newTTradeAssetReservationModel(conn, c, opts...),
	}
}

func (m *customTTradeAssetReservationModel) FindOneByReservationNoForUpdate(ctx context.Context, tenantID int64, reservationNo string) (*TTradeAssetReservation, error) {
	var item TTradeAssetReservation
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND reservation_no = ? LIMIT 1 FOR UPDATE", tTradeAssetReservationRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantID, reservationNo); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTTradeAssetReservationModel) BeginRelease(ctx context.Context, id int64, updateTimes int64) (bool, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeAssetReservationIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeAssetReservationTenantIdReservationNoPrefix, item.TenantId, item.ReservationNo)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET status = 5, last_error_msg = '', next_retry_at = 0, version = version + 1, update_times = ? WHERE id = ? AND status IN (2, 3, 5, 7) AND consumed_amount + released_amount < reserved_amount", m.table)
		return conn.ExecCtx(ctx, query, updateTimes, id)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *customTTradeAssetReservationModel) MarkSettlementFailure(ctx context.Context, id, retryStatus int64, terminal bool, nextRetryAt int64, message string, updateTimes int64) error {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	status := retryStatus
	if terminal {
		status = 7
	}
	item.Status = status
	item.RetryCount++
	item.NextRetryAt = nextRetryAt
	item.LastErrorMsg = message
	item.UpdateTimes = updateTimes
	item.Version++
	return m.Update(ctx, item)
}

func (m *customTTradeAssetReservationModel) AddConsumed(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error) {
	return m.addSettledAmount(ctx, id, "consumed_amount", amount, updateTimes)
}

func (m *customTTradeAssetReservationModel) AddReleased(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error) {
	return m.addSettledAmount(ctx, id, "released_amount", amount, updateTimes)
}

func (m *customTTradeAssetReservationModel) addSettledAmount(ctx context.Context, id int64, column string, amount decimal.Decimal, updateTimes int64) (bool, error) {
	if id <= 0 || !amount.IsPositive() || (column != "consumed_amount" && column != "released_amount") {
		return false, nil
	}
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeAssetReservationIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeAssetReservationTenantIdReservationNoPrefix, item.TenantId, item.ReservationNo)
	statusSQL := "CASE WHEN consumed_amount + ? = reserved_amount THEN 4 WHEN consumed_amount + ? + released_amount = reserved_amount THEN 6 ELSE 3 END"
	if column == "released_amount" {
		statusSQL = "CASE WHEN consumed_amount + released_amount + ? = reserved_amount THEN 6 ELSE 5 END"
	}
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		// Calculate the status before assigning the settled amount. MySQL
		// evaluates single-table UPDATE assignments from left to right; putting
		// the amount assignment first makes statusSQL observe the new value and
		// add the current amount a second time.
		query := fmt.Sprintf("UPDATE %s SET status = %s, %s = %s + ?, next_retry_at = 0, last_error_msg = '', version = version + 1, update_times = ? WHERE id = ? AND consumed_amount + released_amount + ? <= reserved_amount", m.table, statusSQL, column, column)
		if column == "consumed_amount" {
			return conn.ExecCtx(ctx, query, amount, amount, amount, updateTimes, id, amount)
		}
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, id, amount)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *customTTradeAssetReservationModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TTradeAssetReservation, int64, error) {
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
