package models

import (
	"context"
	"database/sql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TTradeMarketSnapshot struct {
	Id                int64           `db:"id"`
	TenantId          int64           `db:"tenant_id"`
	SnapshotId        string          `db:"snapshot_id"`
	SnapshotKind      string          `db:"snapshot_kind"`
	SymbolId          int64           `db:"symbol_id"`
	Source            string          `db:"source"`
	Price             decimal.Decimal `db:"price"`
	MarkPrice         decimal.Decimal `db:"mark_price"`
	IndexPrice        decimal.Decimal `db:"index_price"`
	FundingRate       decimal.Decimal `db:"funding_rate"`
	SourceTimestamp   int64           `db:"source_timestamp"`
	SnapshotTimestamp int64           `db:"snapshot_timestamp"`
	Revision          int64           `db:"revision"`
	FormulaVersion    string          `db:"formula_version"`
	Confirmed         int64           `db:"confirmed"`
	RawPayload        string          `db:"raw_payload"`
	CreateTimes       int64           `db:"create_times"`
}
type TTradeMarketSnapshotModel interface {
	InsertIgnore(ctx context.Context, row *TTradeMarketSnapshot) (sql.Result, error)
	FindOneBySnapshotID(ctx context.Context, id string) (*TTradeMarketSnapshot, error)
	FindPage(ctx context.Context, tenantID, symbolID, cursor, limit, start, end int64, kind string) ([]*TTradeMarketSnapshot, int64, error)
}
type tradeMarketSnapshotModel struct{ conn sqlx.SqlConn }

func NewTTradeMarketSnapshotModel(conn sqlx.SqlConn) TTradeMarketSnapshotModel {
	return &tradeMarketSnapshotModel{conn: conn}
}
func (m *tradeMarketSnapshotModel) InsertIgnore(ctx context.Context, r *TTradeMarketSnapshot) (sql.Result, error) {
	return m.conn.ExecCtx(ctx, "INSERT IGNORE INTO t_trade_market_snapshot(tenant_id,snapshot_id,snapshot_kind,symbol_id,source,price,mark_price,index_price,funding_rate,source_timestamp,snapshot_timestamp,revision,formula_version,confirmed,raw_payload,create_times) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", r.TenantId, r.SnapshotId, r.SnapshotKind, r.SymbolId, r.Source, r.Price, r.MarkPrice, r.IndexPrice, r.FundingRate, r.SourceTimestamp, r.SnapshotTimestamp, r.Revision, r.FormulaVersion, r.Confirmed, r.RawPayload, r.CreateTimes)
}
func (m *tradeMarketSnapshotModel) FindOneBySnapshotID(ctx context.Context, id string) (*TTradeMarketSnapshot, error) {
	var r TTradeMarketSnapshot
	err := m.conn.QueryRowCtx(ctx, &r, "SELECT id,tenant_id,snapshot_id,snapshot_kind,symbol_id,source,price,mark_price,index_price,funding_rate,source_timestamp,snapshot_timestamp,revision,formula_version,confirmed,raw_payload,create_times FROM t_trade_market_snapshot WHERE snapshot_id=?", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func (m *tradeMarketSnapshotModel) FindPage(ctx context.Context, t, sym, cursor, limit, start, end int64, kind string) ([]*TTradeMarketSnapshot, int64, error) {
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
	if err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(1) FROM t_trade_market_snapshot WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	if cursor > 0 {
		where += " AND id<?"
		args = append(args, cursor)
	}
	args = append(args, limit)
	var rows []*TTradeMarketSnapshot
	err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT id,tenant_id,snapshot_id,snapshot_kind,symbol_id,source,price,mark_price,index_price,funding_rate,source_timestamp,snapshot_timestamp,revision,formula_version,confirmed,raw_payload,create_times FROM t_trade_market_snapshot WHERE "+where+" ORDER BY id DESC LIMIT ?", args...)
	return rows, total, err
}
