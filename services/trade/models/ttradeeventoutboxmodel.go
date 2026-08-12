package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeEventOutboxModel = (*customTTradeEventOutboxModel)(nil)

type (
	TradeEventOutboxPageFilter struct {
		TenantId    int64
		EventType   string
		BizType     string
		BizId       string
		EventStatus int64
		TimeStart   int64
		TimeEnd     int64
	}

	// TTradeEventOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeEventOutboxModel.
	TTradeEventOutboxModel interface {
		tTradeEventOutboxModel
		FindPage(ctx context.Context, filter TradeEventOutboxPageFilter, cursor int64, limit int64, knownCounts ...int64) ([]*TTradeEventOutbox, int64, error)
		FindDispatchable(ctx context.Context, tenantID, now, staleBefore, cursor, limit int64, eventTypes []string) ([]*TTradeEventOutbox, error)
		ClaimDispatch(ctx context.Context, id int64, claimant string, now, staleBefore int64) (bool, error)
		MarkDelivered(ctx context.Context, id int64, claimant string, now int64) (bool, error)
		MarkDeliveryFailed(ctx context.Context, id int64, claimant string, now, nextRetryAt int64, errorMessage string) (bool, error)
		ResetForManualRetry(ctx context.Context, id, operatorID, now int64) (bool, error)
	}

	customTTradeEventOutboxModel struct {
		*defaultTTradeEventOutboxModel
	}
)

// NewTTradeEventOutboxModel returns a model for the database table.
func NewTTradeEventOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeEventOutboxModel {
	return &customTTradeEventOutboxModel{
		defaultTTradeEventOutboxModel: newTTradeEventOutboxModel(conn, c, opts...),
	}
}

// Insert applies defaults that cannot be left to MySQL because the generated
// model explicitly includes every column in the INSERT statement.
func (m *customTTradeEventOutboxModel) Insert(ctx context.Context, data *TTradeEventOutbox) (sql.Result, error) {
	applyTradeEventOutboxDefaults(data)
	return m.defaultTTradeEventOutboxModel.Insert(ctx, data)
}

func applyTradeEventOutboxDefaults(data *TTradeEventOutbox) {
	if data == nil {
		return
	}
	if data.PayloadVersion <= 0 {
		data.PayloadVersion = 1
	}
}

func (m *customTTradeEventOutboxModel) FindDispatchable(ctx context.Context, tenantID, now, staleBefore, cursor, limit int64, eventTypes []string) ([]*TTradeEventOutbox, error) {
	limit = sqlutil.NormalizeLimit(limit)
	tenantFilter := ""
	args := make([]any, 0, len(eventTypes)+5)
	if tenantID > 0 {
		tenantFilter = "tenant_id = ? AND "
		args = append(args, tenantID)
	}
	eventFilter := ""
	if len(eventTypes) > 0 {
		marks := strings.TrimSuffix(strings.Repeat("?,", len(eventTypes)), ",")
		eventFilter = fmt.Sprintf(" AND event_type IN (%s)", marks)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s1 = 1%s AND id > ? AND (((event_status = 1 OR (event_status = 3 AND next_retry_at <= ?)) AND (max_retry_count = 0 OR retry_count < max_retry_count)) OR (event_status = 5 AND claimed_at <= ?)) ORDER BY id ASC LIMIT ?", tTradeEventOutboxRows, m.table, tenantFilter, eventFilter)
	for _, eventType := range eventTypes {
		args = append(args, eventType)
	}
	args = append(args, cursor, now, staleBefore, limit)
	var list []*TTradeEventOutbox
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTTradeEventOutboxModel) ClaimDispatch(ctx context.Context, id int64, claimant string, now, staleBefore int64) (bool, error) {
	return m.conditionalEventUpdate(ctx, id, "event_status = 5, retry_count = retry_count + 1, claimed_by = ?, claimed_at = ?, update_times = ?", []any{claimant, now, now}, "(event_status = 1 OR (event_status = 3 AND next_retry_at <= ?) OR (event_status = 5 AND claimed_at <= ?)) AND (max_retry_count = 0 OR retry_count < max_retry_count)", now, staleBefore)
}

func (m *customTTradeEventOutboxModel) MarkDelivered(ctx context.Context, id int64, claimant string, now int64) (bool, error) {
	return m.conditionalEventUpdate(ctx, id, "event_status = 2, delivered_at = ?, claimed_by = '', claimed_at = 0, next_retry_at = 0, last_error_msg = '', update_times = ?", []any{now, now}, "event_status = 5 AND claimed_by = ?", claimant)
}

func (m *customTTradeEventOutboxModel) MarkDeliveryFailed(ctx context.Context, id int64, claimant string, now, nextRetryAt int64, errorMessage string) (bool, error) {
	errorMessage = truncateTradeEventError(errorMessage)
	return m.conditionalEventUpdate(ctx, id, "event_status = IF(max_retry_count > 0 AND retry_count >= max_retry_count, 6, 3), next_retry_at = ?, last_error_msg = ?, claimed_by = '', claimed_at = 0, update_times = ?", []any{nextRetryAt, errorMessage, now}, "event_status = 5 AND claimed_by = ?", claimant)
}

func truncateTradeEventError(message string) string {
	runes := []rune(message)
	if len(runes) <= 500 {
		return message
	}
	return string(runes[:500])
}

func (m *customTTradeEventOutboxModel) ResetForManualRetry(ctx context.Context, id, operatorID, now int64) (bool, error) {
	return m.conditionalEventUpdate(ctx, id, "event_status = 1, retry_count = 0, next_retry_at = ?, last_error_msg = '', claimed_by = '', claimed_at = 0, delivered_at = 0, operator_id = ?, update_times = ?", []any{now, operatorID, now}, "event_status IN (3, 6)")
}

func (m *customTTradeEventOutboxModel) conditionalEventUpdate(ctx context.Context, id int64, setClause string, setArgs []any, whereClause string, whereArgs ...any) (bool, error) {
	item, err := m.FindOne(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTTradeEventOutboxIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v", cacheTTradeEventOutboxTenantIdEventNoPrefix, item.TenantId, item.EventNo)
	args := append(append([]any{}, setArgs...), id)
	args = append(args, whereArgs...)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND %s", m.table, setClause, whereClause)
		return conn.ExecCtx(ctx, query, args...)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTTradeEventOutboxModel) FindPage(ctx context.Context, filter TradeEventOutboxPageFilter, cursor int64, limit int64, knownCounts ...int64) ([]*TTradeEventOutbox, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("event_type", filter.EventType)
	builder.EqString("biz_type", filter.BizType)
	builder.EqString("biz_id", filter.BizId)
	builder.EqInt64("event_status", filter.EventStatus)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)

	where := builder.Where()
	args := builder.Args()

	total := sqlutil.KnownCount(knownCounts...)
	if total <= 0 {
		countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
		if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
			return nil, 0, err
		}
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeEventOutboxRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeEventOutbox
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
