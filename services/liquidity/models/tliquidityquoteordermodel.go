package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteOrderModel = (*customTLiquidityQuoteOrderModel)(nil)

type (
	LiquidityQuoteOrderPageFilter struct {
		ConfigId, ProviderId, SymbolId int64
		Side, Status                   int64
		Keyword                        string
		TimeStart, TimeEnd             int64
	}
	// TLiquidityQuoteOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteOrderModel.
	TLiquidityQuoteOrderModel interface {
		tLiquidityQuoteOrderModel
		FindPage(ctx context.Context, filter LiquidityQuoteOrderPageFilter, cursor, limit int64, knownCounts ...int64) ([]*TLiquidityQuoteOrder, int64, error)
		FindByInternalIdentity(ctx context.Context, internalOrderID int64, internalOrderNo, clientOrderID string) (*TLiquidityQuoteOrder, error)
		CancelActiveByConfig(ctx context.Context, configID int64, reason string, now, pendingStatus, canceledStatus, cancelingStatus int64) error
		FindActiveByConfig(ctx context.Context, configID int64) ([]*TLiquidityQuoteOrder, error)
		HasUncertainByConfig(ctx context.Context, configID int64) (bool, error)
		FindRecoveryCandidates(ctx context.Context, filter LiquidityQuoteOrderPageFilter, limit int64) ([]*TLiquidityQuoteOrder, error)
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

func (m *customTLiquidityQuoteOrderModel) FindActiveByConfig(ctx context.Context, configID int64) ([]*TLiquidityQuoteOrder, error) {
	var rows []*TLiquidityQuoteOrder
	query := fmt.Sprintf("SELECT %s FROM %s WHERE config_id = ? AND status IN (?, ?, ?, ?) ORDER BY id ASC", tLiquidityQuoteOrderRows, m.table)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query,
		configID,
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN),
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PART_FILLED),
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING),
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customTLiquidityQuoteOrderModel) HasUncertainByConfig(ctx context.Context, configID int64) (bool, error) {
	var exists int64
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE config_id = ? AND status = ? LIMIT 1)", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &exists, query,
		configID,
		int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN),
	); err != nil {
		return false, err
	}
	return exists == 1, nil
}

// FindRecoveryCandidates rotates unresolved orders by their last inspection
// time. This prevents a permanently unresolved newest batch from starving older
// rows on every scheduled recovery run.
func (m *customTLiquidityQuoteOrderModel) FindRecoveryCandidates(ctx context.Context, filter LiquidityQuoteOrderPageFilter, limit int64) ([]*TLiquidityQuoteOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("status", int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN))
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY update_times ASC, id ASC LIMIT ?", tLiquidityQuoteOrderRows, m.table, b.Where())
	args := append(b.Args(), limit)
	var rows []*TLiquidityQuoteOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customTLiquidityQuoteOrderModel) FindByInternalIdentity(ctx context.Context, internalOrderID int64, internalOrderNo, clientOrderID string) (*TLiquidityQuoteOrder, error) {
	b := sqlutil.NewPageQueryBuilder()
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

func (m *customTLiquidityQuoteOrderModel) FindPage(ctx context.Context, filter LiquidityQuoteOrderPageFilter, cursor, limit int64, knownCounts ...int64) ([]*TLiquidityQuoteOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
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
	total := sqlutil.KnownCount(knownCounts...)
	if total <= 0 {
		if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
			return nil, 0, err
		}
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

func (m *customTLiquidityQuoteOrderModel) CancelActiveByConfig(ctx context.Context, configID int64, reason string, now, pendingStatus, canceledStatus, cancelingStatus int64) error {
	var rows []*TLiquidityQuoteOrder
	query := fmt.Sprintf("SELECT %s FROM %s WHERE config_id = ? AND status IN (?, 2, 3)", tLiquidityQuoteOrderRows, m.table)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, configID, pendingStatus); err != nil {
		return err
	}
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		update := fmt.Sprintf(`UPDATE %s SET status = CASE WHEN status = ? THEN ? ELSE ? END,
			cancel_reason = ?, version = version + 1, update_times = ?
			WHERE config_id = ? AND status IN (?, 2, 3)`, m.table)
		return conn.ExecCtx(ctx, update, pendingStatus, canceledStatus, cancelingStatus, reason, now, configID, pendingStatus)
	}, quoteOrderCacheKeys(rows)...)
	return err
}

func quoteOrderCacheKeys(rows []*TLiquidityQuoteOrder) []string {
	keys := make([]string, 0, len(rows)*3)
	for _, row := range rows {
		keys = append(keys,
			fmt.Sprintf("%s%v", cacheTLiquidityQuoteOrderIdPrefix, row.Id),
			fmt.Sprintf("%s%v", cacheTLiquidityQuoteOrderClientOrderIdPrefix, row.ClientOrderId),
			fmt.Sprintf("%s%v", cacheTLiquidityQuoteOrderQuoteNoPrefix, row.QuoteNo))
	}
	return keys
}
