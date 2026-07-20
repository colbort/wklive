package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeEventInboxModel = (*customTTradeEventInboxModel)(nil)

type (
	// TTradeEventInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeEventInboxModel.
	TTradeEventInboxModel interface {
		tTradeEventInboxModel
		Claim(ctx context.Context, consumer string, tenantID int64, eventNo, eventType string, now, staleBefore int64) (claimed, completed bool, lease int64, err error)
		Complete(ctx context.Context, consumer string, tenantID int64, eventNo string, lease, now int64) (bool, error)
		Fail(ctx context.Context, consumer string, tenantID int64, eventNo string, lease int64, errorMessage string, now int64) error
	}

	customTTradeEventInboxModel struct {
		*defaultTTradeEventInboxModel
	}
)

// NewTTradeEventInboxModel returns a model for the database table.
func NewTTradeEventInboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeEventInboxModel {
	return &customTTradeEventInboxModel{
		defaultTTradeEventInboxModel: newTTradeEventInboxModel(conn, c, opts...),
	}
}

func (m *defaultTTradeEventInboxModel) Claim(ctx context.Context, consumer string, tenantID int64, eventNo, eventType string, now, staleBefore int64) (bool, bool, int64, error) {
	item, err := m.FindOneByConsumerTenantIdEventNo(ctx, consumer, tenantID, eventNo)
	if errors.Is(err, ErrNotFound) {
		_, insertErr := m.Insert(ctx, &TTradeEventInbox{TenantId: tenantID, Consumer: consumer, EventNo: eventNo, EventType: eventType, Status: 1, CreateTimes: now, UpdateTimes: now})
		if insertErr == nil {
			return true, false, now, nil
		}
		item, err = m.FindOneByConsumerTenantIdEventNo(ctx, consumer, tenantID, eventNo)
		if err != nil {
			return false, false, 0, insertErr
		}
	} else if err != nil {
		return false, false, 0, err
	}
	if item.Status == 2 {
		return false, true, 0, nil
	}
	if item.Status == 1 && item.UpdateTimes > staleBefore {
		return false, false, 0, nil
	}
	changed, err := m.conditionalInboxUpdate(ctx, item, "status = 1, retry_count = retry_count + 1, last_error_msg = '', update_times = ?", []any{now}, "(status = 3 AND update_times <= ?) OR (status = 1 AND update_times <= ?)", now, staleBefore)
	return changed, false, now, err
}

func (m *defaultTTradeEventInboxModel) Complete(ctx context.Context, consumer string, tenantID int64, eventNo string, lease, now int64) (bool, error) {
	item, err := m.FindOneByConsumerTenantIdEventNo(ctx, consumer, tenantID, eventNo)
	if err != nil {
		return false, err
	}
	return m.conditionalInboxUpdate(ctx, item, "status = 2, last_error_msg = '', update_times = ?", []any{now}, "status = 1 AND update_times = ?", lease)
}

func (m *defaultTTradeEventInboxModel) Fail(ctx context.Context, consumer string, tenantID int64, eventNo string, lease int64, errorMessage string, now int64) error {
	errorMessage = truncateTradeEventError(errorMessage)
	item, err := m.FindOneByConsumerTenantIdEventNo(ctx, consumer, tenantID, eventNo)
	if err != nil {
		return err
	}
	nextRetryAt := now + time.Second.Milliseconds()
	_, err = m.conditionalInboxUpdate(ctx, item, "status = 3, last_error_msg = ?, update_times = ?", []any{errorMessage, nextRetryAt}, "status = 1 AND update_times = ?", lease)
	return err
}

func (m *defaultTTradeEventInboxModel) conditionalInboxUpdate(ctx context.Context, item *TTradeEventInbox, setClause string, setArgs []any, whereClause string, whereArgs ...any) (bool, error) {
	idKey := fmt.Sprintf("%s%v", cacheTTradeEventInboxIdPrefix, item.Id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v", cacheTTradeEventInboxConsumerTenantIdEventNoPrefix, item.Consumer, item.TenantId, item.EventNo)
	args := append(append([]any{}, setArgs...), item.Id)
	args = append(args, whereArgs...)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND (%s)", m.table, setClause, whereClause)
		return conn.ExecCtx(ctx, query, args...)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}
