package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderSecondsModel = (*customTTradeOrderSecondsModel)(nil)

const (
	secondsManualRetryThreshold = int64(20)
	secondsMaxRetryShift        = int64(10)
)

type (
	SecondsOrderWorkItem struct {
		TTradeOrderSeconds
		OrderNo     string `db:"order_no"`
		UserId      int64  `db:"user_id"`
		SymbolId    int64  `db:"symbol_id"`
		OrderStatus int64  `db:"order_status"`
	}
	// TTradeOrderSecondsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderSecondsModel.
	TTradeOrderSecondsModel interface {
		tTradeOrderSecondsModel
		FindWork(ctx context.Context, tenantID, status, dueAt, cursor, limit int64) ([]*SecondsOrderWorkItem, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrderSeconds, error)
		SumExposure(ctx context.Context, tenantID, symbolID int64, statuses []int64) (decimal.Decimal, error)
		ClaimSettlement(ctx context.Context, id, now, staleBefore int64) (bool, int64, error)
		ClaimRefund(ctx context.Context, id, now, staleBefore int64) (bool, int64, error)
		ClaimActivation(ctx context.Context, id, now, staleBefore int64) (bool, int64, error)
		MarkWorkFailure(ctx context.Context, id, status, lease int64, message string, now int64) (bool, error)
	}

	customTTradeOrderSecondsModel struct {
		*defaultTTradeOrderSecondsModel
	}
)

func (m *defaultTTradeOrderSecondsModel) ClaimActivation(ctx context.Context, id, now, staleBefore int64) (bool, int64, error) {
	return m.claimWork(ctx, id, int64(1), now, "settlement_status = 1 AND (update_times = 0 OR update_times <= ?)", staleBefore)
}

func (m *defaultTTradeOrderSecondsModel) MarkWorkFailure(ctx context.Context, id, status, lease int64, message string, now int64) (bool, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	retry := item.RetryCount + 1
	nextStatus, nextRetry, manual := secondsRetryState(status, retry, now)
	message = strings.TrimSpace(message)
	messageRunes := []rune(message)
	if len(messageRunes) > 500 {
		message = string(messageRunes[:500])
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeOrderSecondsIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeOrderSecondsTenantIdOrderIdPrefix, item.TenantId, item.OrderId)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET settlement_status=?,retry_count=?,next_retry_at=?,last_error_msg=?,version=version+1,update_times=0 WHERE id=? AND settlement_status=? AND update_times=?", m.table)
		return conn.ExecCtx(ctx, query, nextStatus, retry, nextRetry, message, id, status, lease)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected != 1 {
		return false, sql.ErrNoRows
	}
	return manual, nil
}

func secondsRetryState(status, retry, now int64) (nextStatus, nextRetry int64, manual bool) {
	if retry >= secondsManualRetryThreshold {
		return int64(7), 0, true
	}
	shift := retry
	if shift > secondsMaxRetryShift {
		shift = secondsMaxRetryShift
	}
	return status, now + (int64(1)<<shift)*int64(time.Second/time.Millisecond), false
}

// NewTTradeOrderSecondsModel returns a model for the database table.
func NewTTradeOrderSecondsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderSecondsModel {
	return &customTTradeOrderSecondsModel{
		defaultTTradeOrderSecondsModel: newTTradeOrderSecondsModel(conn, c, opts...),
	}
}

func (m *defaultTTradeOrderSecondsModel) ClaimSettlement(ctx context.Context, id, now, staleBefore int64) (bool, int64, error) {
	return m.claimWork(ctx, id, int64(3), now, "settlement_status = 2 OR (settlement_status = 3 AND update_times <= ?)", staleBefore)
}

func (m *defaultTTradeOrderSecondsModel) ClaimRefund(ctx context.Context, id, now, staleBefore int64) (bool, int64, error) {
	return m.claimWork(ctx, id, int64(5), now, "settlement_status = 5 AND update_times <= ?", staleBefore)
}

func (m *defaultTTradeOrderSecondsModel) claimWork(ctx context.Context, id, targetStatus, now int64, condition string, args ...any) (bool, int64, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, 0, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeOrderSecondsIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeOrderSecondsTenantIdOrderIdPrefix, item.TenantId, item.OrderId)
	queryArgs := []any{targetStatus, now, id}
	queryArgs = append(queryArgs, args...)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET settlement_status = ?, version = version + 1, update_times = ? WHERE id = ? AND (%s)", m.table, condition)
		return conn.ExecCtx(ctx, query, queryArgs...)
	}, idKey, uniqueKey)
	if err != nil {
		return false, 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return false, 0, err
	}
	return true, now, nil
}

func (m *defaultTTradeOrderSecondsModel) FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrderSeconds, error) {
	var row TTradeOrderSeconds
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? FOR UPDATE", tTradeOrderSecondsRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultTTradeOrderSecondsModel) FindWork(ctx context.Context, tenantID, status, dueAt, cursor, limit int64) ([]*SecondsOrderWorkItem, error) {
	where := "s.settlement_status = ? AND s.id > ? AND (s.next_retry_at=0 OR s.next_retry_at<=?)"
	args := []any{status, cursor, time.Now().UnixMilli()}
	if tenantID > 0 {
		where += " AND s.tenant_id = ?"
		args = append(args, tenantID)
	}
	if dueAt > 0 {
		where += " AND s.expire_time <= ?"
		args = append(args, dueAt)
	}
	args = append(args, limit)
	query := fmt.Sprintf("SELECT %s, o.order_no, o.user_id, o.symbol_id, o.status AS order_status FROM %s s JOIN t_trade_order o ON o.id=s.order_id AND o.tenant_id=s.tenant_id WHERE %s ORDER BY s.id ASC LIMIT ?", prefixedFields(tTradeOrderSecondsFieldNames, "s"), m.table, where)
	var rows []*SecondsOrderWorkItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *defaultTTradeOrderSecondsModel) SumExposure(ctx context.Context, tenantID, symbolID int64, statuses []int64) (decimal.Decimal, error) {
	if len(statuses) == 0 {
		return decimal.Zero, nil
	}
	holders := "?"
	args := []any{tenantID, symbolID, statuses[0]}
	for _, status := range statuses[1:] {
		holders += ",?"
		args = append(args, status)
	}
	var amount decimal.Decimal
	query := fmt.Sprintf("SELECT COALESCE(SUM(s.stake_amount),0) FROM %s s JOIN t_trade_order o ON o.id=s.order_id WHERE s.tenant_id=? AND o.symbol_id=? AND s.settlement_status IN (%s)", m.table, holders)
	if err := m.QueryRowNoCacheCtx(ctx, &amount, query, args...); err != nil {
		return decimal.Zero, err
	}
	return amount, nil
}

func prefixedFields(fields []string, alias string) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ","
		}
		result += alias + "." + field
	}
	return result
}
