package models

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeMarketSnapshotModel = (*customTTradeMarketSnapshotModel)(nil)

type (
	// TTradeMarketSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeMarketSnapshotModel.
	TTradeMarketSnapshotModel interface {
		tTradeMarketSnapshotModel
		InsertIgnore(ctx context.Context, row *TTradeMarketSnapshot) (sql.Result, error)
		FindOneBySnapshotID(ctx context.Context, id string) (*TTradeMarketSnapshot, error)
		FindLatestConfirmed(ctx context.Context, tenantID, symbolID, minSourceTimestamp int64) (*TTradeMarketSnapshot, error)
		FindPage(ctx context.Context, tenantID, symbolID, cursor, limit, start, end int64, kind string) ([]*TTradeMarketSnapshot, int64, error)
	}

	customTTradeMarketSnapshotModel struct {
		*defaultTTradeMarketSnapshotModel
	}
)

// NewTTradeMarketSnapshotModel returns a model for the database table.
func NewTTradeMarketSnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeMarketSnapshotModel {
	return &customTTradeMarketSnapshotModel{
		defaultTTradeMarketSnapshotModel: newTTradeMarketSnapshotModel(conn, c, opts...),
	}
}

func (m *defaultTTradeMarketSnapshotModel) InsertIgnore(ctx context.Context, r *TTradeMarketSnapshot) (sql.Result, error) {
	return m.ExecNoCacheCtx(ctx, "INSERT IGNORE INTO t_trade_market_snapshot(tenant_id,snapshot_id,snapshot_kind,symbol_id,source,price,mark_price,index_price,funding_rate,source_timestamp,snapshot_timestamp,revision,formula_version,confirmed,raw_payload,create_times) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", r.TenantId, r.SnapshotId, r.SnapshotKind, r.SymbolId, r.Source, r.Price, r.MarkPrice, r.IndexPrice, r.FundingRate, r.SourceTimestamp, r.SnapshotTimestamp, r.Revision, r.FormulaVersion, r.Confirmed, r.RawPayload, r.CreateTimes)
}
func (m *defaultTTradeMarketSnapshotModel) FindOneBySnapshotID(ctx context.Context, id string) (*TTradeMarketSnapshot, error) {
	var r TTradeMarketSnapshot
	err := m.QueryRowNoCacheCtx(ctx, &r, "SELECT "+tTradeMarketSnapshotRows+" FROM t_trade_market_snapshot WHERE snapshot_id=?", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (m *defaultTTradeMarketSnapshotModel) FindLatestConfirmed(ctx context.Context, tenantID, symbolID, minSourceTimestamp int64) (*TTradeMarketSnapshot, error) {
	var row TTradeMarketSnapshot
	err := m.QueryRowNoCacheCtx(ctx, &row, "SELECT "+tTradeMarketSnapshotRows+" FROM t_trade_market_snapshot WHERE tenant_id IN (?,0) AND symbol_id=? AND confirmed=1 AND source_timestamp>=? ORDER BY (tenant_id=?) DESC, source_timestamp DESC, revision DESC, id DESC LIMIT 1", tenantID, symbolID, minSourceTimestamp, tenantID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultTTradeMarketSnapshotModel) FindPage(ctx context.Context, t, sym, cursor, limit, start, end int64, kind string) ([]*TTradeMarketSnapshot, int64, error) {
	where := "tenant_id=?"
	args := []any{t}
	if sym > 0 {
		where += " AND symbol_id=?"
		args = append(args, sym)
	}
	if kind != "" {
		where += " AND snapshot_kind=?"
		args = append(args, kind)
	}
	if start > 0 {
		where += " AND source_timestamp>=?"
		args = append(args, start)
	}
	if end > 0 {
		where += " AND source_timestamp<=?"
		args = append(args, end)
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(1) FROM t_trade_market_snapshot WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	if cursor > 0 {
		where += " AND id<?"
		args = append(args, cursor)
	}
	args = append(args, limit)
	var rows []*TTradeMarketSnapshot
	err := m.QueryRowsNoCacheCtx(ctx, &rows, "SELECT "+tTradeMarketSnapshotRows+" FROM t_trade_market_snapshot WHERE "+where+" ORDER BY id DESC LIMIT ?", args...)
	return rows, total, err
}
