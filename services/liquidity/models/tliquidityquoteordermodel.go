package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteOrderModel = (*customTLiquidityQuoteOrderModel)(nil)

type (
	LiquidityQuoteOrderPageFilter struct {
		TenantId, ConfigId, ProviderId, SymbolId int64
		Side, Status                             int64
		Keyword                                  string
		TimeStart, TimeEnd                       int64
	}
	// TLiquidityQuoteOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteOrderModel.
	TLiquidityQuoteOrderModel interface {
		tLiquidityQuoteOrderModel
		FindPage(ctx context.Context, filter LiquidityQuoteOrderPageFilter, cursor, limit int64) ([]*TLiquidityQuoteOrder, int64, error)
		FindByInternalIdentity(ctx context.Context, tenantID, internalOrderID int64, internalOrderNo, clientOrderID string) (*TLiquidityQuoteOrder, error)
		CancelActiveByConfig(ctx context.Context, tenantID, configID int64, reason string, now, pendingStatus, canceledStatus, cancelingStatus int64) error
	}

	customTLiquidityQuoteOrderModel struct {
		*defaultTLiquidityQuoteOrderModel
	}
)

// NewTLiquidityQuoteOrderModel returns a model for the database table.
func NewTLiquidityQuoteOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityQuoteOrderModel {
	return &customTLiquidityQuoteOrderModel{
		defaultTLiquidityQuoteOrderModel: newTLiquidityQuoteOrderModel(conn, c, opts...),
	}
}

func (m *customTLiquidityQuoteOrderModel) FindByInternalIdentity(ctx context.Context, tenantID, internalOrderID int64, internalOrderNo, clientOrderID string) (*TLiquidityQuoteOrder, error) {
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", tenantID)
	switch {
	case internalOrderID > 0:
		b.EqInt64("internal_order_id", internalOrderID)
	case internalOrderNo != "":
		b.And("internal_order_no = ?", internalOrderNo)
	case clientOrderID != "":
		b.And("client_order_id = ?", clientOrderID)
	default:
		return nil, ErrNotFound
	}
	var row TLiquidityQuoteOrder
	err := m.QueryRowNoCacheCtx(ctx, &row, fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 1", tLiquidityQuoteOrderRows, m.table, b.Where()), b.Args()...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquidityQuoteOrderModel) FindPage(ctx context.Context, filter LiquidityQuoteOrderPageFilter, cursor, limit int64) ([]*TLiquidityQuoteOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("side", filter.Side)
	b.EqInt64("status", filter.Status)
	b.GteInt64("create_times", filter.TimeStart)
	b.LteInt64("create_times", filter.TimeEnd)
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		b.And("(quote_no LIKE ? OR internal_order_no LIKE ? OR client_order_id LIKE ?)", kw, kw, kw)
	}
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityQuoteOrderRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityQuoteOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *customTLiquidityQuoteOrderModel) CancelActiveByConfig(ctx context.Context, tenantID, configID int64, reason string, now, pendingStatus, canceledStatus, cancelingStatus int64) error {
	var rows []*TLiquidityQuoteOrder
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND config_id = ? AND status IN (?, 2, 3)", tLiquidityQuoteOrderRows, m.table)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, tenantID, configID, pendingStatus); err != nil {
		return err
	}
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		update := fmt.Sprintf(`UPDATE %s SET status = CASE WHEN status = ? THEN ? ELSE ? END,
			cancel_reason = ?, version = version + 1, update_times = ?
			WHERE tenant_id = ? AND config_id = ? AND status IN (?, 2, 3)`, m.table)
		return conn.ExecCtx(ctx, update, pendingStatus, canceledStatus, cancelingStatus, reason, now, tenantID, configID, pendingStatus)
	}, quoteOrderCacheKeys(rows)...)
	return err
}

func quoteOrderCacheKeys(rows []*TLiquidityQuoteOrder) []string {
	keys := make([]string, 0, len(rows)*3)
	for _, row := range rows {
		keys = append(keys,
			fmt.Sprintf("%s%v", cacheTLiquidityQuoteOrderIdPrefix, row.Id),
			fmt.Sprintf("%s%v:%v", cacheTLiquidityQuoteOrderTenantIdClientOrderIdPrefix, row.TenantId, row.ClientOrderId),
			fmt.Sprintf("%s%v:%v", cacheTLiquidityQuoteOrderTenantIdQuoteNoPrefix, row.TenantId, row.QuoteNo))
	}
	return keys
}
