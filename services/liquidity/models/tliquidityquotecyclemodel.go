package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/common"
	"wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityQuoteCycleModel = (*customTLiquidityQuoteCycleModel)(nil)

type (
	LiquidityQuoteCyclePageFilter struct {
		ConfigId, SymbolId, Status int64
		TimeStart, TimeEnd         int64
	}
	// TLiquidityQuoteCycleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityQuoteCycleModel.
	TLiquidityQuoteCycleModel interface {
		tLiquidityQuoteCycleModel
		FindPage(ctx context.Context, filter LiquidityQuoteCyclePageFilter, cursor, limit int64, knownCounts ...int64) ([]*TLiquidityQuoteCycle, int64, error)
		FindLatestByConfig(ctx context.Context, configID int64) (*TLiquidityQuoteCycle, error)
		RefreshExecutionResults(ctx context.Context, configID, now int64) error
	}

	customTLiquidityQuoteCycleModel struct {
		*defaultTLiquidityQuoteCycleModel
	}
)

// NewTLiquidityQuoteCycleModel returns a model for the database table.
func NewTLiquidityQuoteCycleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityQuoteCycleModel {
	return &customTLiquidityQuoteCycleModel{
		defaultTLiquidityQuoteCycleModel: newTLiquidityQuoteCycleModel(conn, c, opts...),
	}
}

func (m *customTLiquidityQuoteCycleModel) RefreshExecutionResults(ctx context.Context, configID, now int64) error {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = ?", tLiquidityQuoteCycleRows, m.table)
	args := []any{int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_EXECUTING)}
	if configID > 0 {
		query += " AND config_id = ?"
		args = append(args, configID)
	}
	var cycles []*TLiquidityQuoteCycle
	if err := m.QueryRowsNoCacheCtx(ctx, &cycles, query, args...); err != nil {
		return err
	}
	type executionCounts struct {
		Pending int64 `db:"pending"`
		Bids    int64 `db:"bids"`
		Asks    int64 `db:"asks"`
		Failed  int64 `db:"failed"`
	}
	for _, cycle := range cycles {
		var counts executionCounts
		countQuery := `SELECT
			COALESCE(SUM(status IN (?, ?, ?)), 0) AS pending,
			COALESCE(SUM(side = ? AND status IN (?, ?, ?)), 0) AS bids,
			COALESCE(SUM(side = ? AND status IN (?, ?, ?)), 0) AS asks,
			COALESCE(SUM(status = ?), 0) AS failed
			FROM t_liquidity_quote_order WHERE cycle_id = ?`
		if err := m.QueryRowNoCacheCtx(ctx, &counts, countQuery,
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN),
			int64(common.Side_SIDE_BUY),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PART_FILLED),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FILLED),
			int64(common.Side_SIDE_SELL),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PART_FILLED),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FILLED),
			int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED),
			cycle.Id,
		); err != nil {
			return err
		}
		cycle.PlacedBidCount, cycle.PlacedAskCount = counts.Bids, counts.Asks
		if counts.Pending > 0 {
			cycle.UpdateTimes = now
			if err := m.Update(ctx, cycle); err != nil {
				return err
			}
			continue
		}
		switch {
		case counts.Bids == cycle.TargetBidCount && counts.Asks == cycle.TargetAskCount:
			cycle.Status = int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_SUCCESS)
		case counts.Bids+counts.Asks > 0:
			cycle.Status = int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_PARTIAL_SUCCESS)
		default:
			cycle.Status = int64(liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_FAILED)
		}
		if counts.Failed > 0 {
			cycle.LastErrorMsg = fmt.Sprintf("%d quote order(s) failed", counts.Failed)
		}
		cycle.FinishedAt, cycle.UpdateTimes = now, now
		if err := m.Update(ctx, cycle); err != nil {
			return err
		}
	}
	return nil
}

func (m *customTLiquidityQuoteCycleModel) FindLatestByConfig(ctx context.Context, configID int64) (*TLiquidityQuoteCycle, error) {
	var row TLiquidityQuoteCycle
	query := fmt.Sprintf("SELECT %s FROM %s WHERE config_id = ? ORDER BY id DESC LIMIT 1", tLiquidityQuoteCycleRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, configID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customTLiquidityQuoteCycleModel) FindPage(ctx context.Context, filter LiquidityQuoteCyclePageFilter, cursor, limit int64, knownCounts ...int64) ([]*TLiquidityQuoteCycle, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("status", filter.Status)
	b.GteInt64("create_times", filter.TimeStart)
	b.LteInt64("create_times", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	total := sqlutil.KnownCount(knownCounts...)
	if total <= 0 {
		if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
			return nil, 0, err
		}
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityQuoteCycleRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityQuoteCycle
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
