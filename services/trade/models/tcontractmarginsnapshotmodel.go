package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ TContractMarginSnapshotModel = (*customTContractMarginSnapshotModel)(nil)

type (
	CrossMarginAggregate struct {
		TenantID           int64           `db:"tenant_id"`
		UserID             int64           `db:"user_id"`
		MarginAsset        string          `db:"margin_asset"`
		PositionMargin     decimal.Decimal `db:"position_margin"`
		MaintenanceMargin  decimal.Decimal `db:"maintenance_margin"`
		UnrealizedPnl      decimal.Decimal `db:"unrealized_pnl"`
		RealizedPnl        decimal.Decimal `db:"realized_pnl"`
		PositionCount      int64           `db:"position_count"`
		PositionVersionSum int64           `db:"position_version_sum"`
		PositionUpdateTime int64           `db:"position_update_time"`
		OrderMargin        decimal.Decimal `db:"order_margin"`
		OrderVersionSum    int64           `db:"order_version_sum"`
		OrderUpdateTime    int64           `db:"order_update_time"`
	}

	CrossMarginAggregateCursor struct {
		TenantID    int64
		UserID      int64
		MarginAsset string
		Valid       bool
	}

	// TContractMarginSnapshotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractMarginSnapshotModel.
	TContractMarginSnapshotModel interface {
		tContractMarginSnapshotModel
		FindPage(ctx context.Context, tenantId, userId int64, marginAsset string, cursor, limit int64) ([]*TContractMarginSnapshot, int64, error)
		UpsertRiskProjection(ctx context.Context, data *TContractMarginSnapshot) (bool, error)
		FindRiskProjectionAggregates(ctx context.Context, tenantID int64, cursor CrossMarginAggregateCursor, limit int) ([]*CrossMarginAggregate, error)
	}

	customTContractMarginSnapshotModel struct {
		*defaultTContractMarginSnapshotModel
	}
)

func (m *defaultTContractMarginSnapshotModel) FindRiskProjectionAggregates(
	ctx context.Context, tenantID int64, cursor CrossMarginAggregateCursor, limit int,
) ([]*CrossMarginAggregate, error) {
	where := "1=1"
	args := make([]any, 0, 10)
	if tenantID > 0 {
		where += " AND u.tenant_id=?"
		args = append(args, tenantID)
	}
	if cursor.Valid {
		where += ` AND (
  u.tenant_id>? OR
  (u.tenant_id=? AND u.user_id>?) OR
  (u.tenant_id=? AND u.user_id=? AND u.margin_asset>?)
)`
		args = append(args,
			cursor.TenantID,
			cursor.TenantID, cursor.UserID,
			cursor.TenantID, cursor.UserID, cursor.MarginAsset,
		)
	}
	args = append(args, limit)
	query := `
WITH
p AS (
  SELECT tenant_id,user_id,margin_asset,
    COALESCE(SUM(position_margin),0) AS position_margin,
    COALESCE(SUM(maintenance_margin),0) AS maintenance_margin,
    COALESCE(SUM(unrealized_pnl),0) AS unrealized_pnl,
    COALESCE(SUM(realized_pnl),0) AS realized_pnl,
    COUNT(1) AS position_count,
    COALESCE(SUM(version),0) AS position_version_sum,
    COALESCE(MAX(update_times),0) AS position_update_time
  FROM t_contract_position
  WHERE margin_mode=1 AND status=1 AND qty>0
  GROUP BY tenant_id,user_id,margin_asset
),
o AS (
  SELECT r.tenant_id,o.user_id,c.margin_asset,
    COALESCE(SUM(GREATEST(r.reserved_amount-r.consumed_amount-r.released_amount,0)),0) AS order_margin,
    COALESCE(SUM(r.version),0) AS order_version_sum,
    COALESCE(MAX(GREATEST(r.update_times,o.update_times,c.update_times)),0) AS order_update_time
  FROM t_trade_asset_reservation r
  JOIN t_trade_order_contract c
    ON c.tenant_id=r.tenant_id AND c.order_id=r.order_id
  JOIN t_trade_order o
    ON o.tenant_id=r.tenant_id AND o.id=r.order_id
  WHERE c.margin_mode=1
    AND r.status IN (1,2,3,5,7)
    AND r.reserved_amount>r.consumed_amount+r.released_amount
  GROUP BY r.tenant_id,o.user_id,c.margin_asset
),
u AS (
  SELECT tenant_id,user_id,margin_asset FROM p
  UNION
  SELECT tenant_id,user_id,margin_asset FROM o
  UNION
  SELECT s.tenant_id,s.user_id,s.margin_asset
  FROM t_contract_margin_snapshot s
  WHERE s.position_count>0 OR s.order_margin>0
)
SELECT u.tenant_id,u.user_id,u.margin_asset,
  COALESCE(p.position_margin,0) AS position_margin,
  COALESCE(p.maintenance_margin,0) AS maintenance_margin,
  COALESCE(p.unrealized_pnl,0) AS unrealized_pnl,
  COALESCE(p.realized_pnl,0) AS realized_pnl,
  COALESCE(p.position_count,0) AS position_count,
  COALESCE(p.position_version_sum,0) AS position_version_sum,
  COALESCE(p.position_update_time,0) AS position_update_time,
  COALESCE(o.order_margin,0) AS order_margin,
  COALESCE(o.order_version_sum,0) AS order_version_sum,
  COALESCE(o.order_update_time,0) AS order_update_time
FROM u
LEFT JOIN p ON p.tenant_id=u.tenant_id AND p.user_id=u.user_id AND p.margin_asset=u.margin_asset
LEFT JOIN o ON o.tenant_id=u.tenant_id AND o.user_id=u.user_id AND o.margin_asset=u.margin_asset
WHERE ` + where + `
ORDER BY u.tenant_id,u.user_id,u.margin_asset
LIMIT ?`
	var groups []*CrossMarginAggregate
	if err := m.QueryRowsNoCacheCtx(ctx, &groups, query, args...); err != nil {
		return nil, err
	}
	return groups, nil
}

// UpsertRiskProjection persists one account/settlement-asset projection and
// only advances its version when a risk input or result changed. ProcessPositions
// is protected by a distributed task lock; the version predicate still prevents
// a manual projection repair from being overwritten by a stale scan.
func (m *defaultTContractMarginSnapshotModel) UpsertRiskProjection(ctx context.Context, data *TContractMarginSnapshot) (bool, error) {
	if data == nil {
		return false, fmt.Errorf("cross margin risk projection is nil")
	}
	var current TContractMarginSnapshot
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND user_id=? AND margin_asset=? LIMIT 1", tContractMarginSnapshotRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &current, query, data.TenantId, data.UserId, data.MarginAsset)
	if err == sqlx.ErrNotFound {
		data.Version = 1
		_, err = m.Insert(ctx, data)
		return err == nil, err
	}
	if err != nil {
		return false, err
	}
	if marginRiskProjectionEqual(&current, data) {
		return false, nil
	}
	data.Id = current.Id
	data.Version = current.Version + 1
	data.CreateTimes = current.CreateTimes
	idKey := fmt.Sprintf("%s%v", cacheTContractMarginSnapshotIdPrefix, current.Id)
	accountKey := fmt.Sprintf("%s%v:%v:%v", cacheTContractMarginSnapshotTenantIdUserIdMarginAssetPrefix, current.TenantId, current.UserId, current.MarginAsset)
	oldSourceKey := fmt.Sprintf("%s%v:%v", cacheTContractMarginSnapshotTenantIdSourceEventNoPrefix, current.TenantId, current.SourceEventNo)
	newSourceKey := fmt.Sprintf("%s%v:%v", cacheTContractMarginSnapshotTenantIdSourceEventNoPrefix, data.TenantId, data.SourceEventNo)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		update := fmt.Sprintf(`UPDATE %s SET
			wallet_balance=?,available_balance=?,frozen_balance=?,
			position_margin=?,order_margin=?,maintenance_margin=?,
			account_equity=?,available_margin=?,risk_rate=?,position_count=?,asset_version=?,
			unrealized_pnl=?,realized_pnl=?,source_event_no=?,snapshot_time=?,
			version=?,update_times=?
			WHERE id=? AND version=?`, m.table)
		return conn.ExecCtx(ctx, update,
			data.WalletBalance, data.AvailableBalance, data.FrozenBalance,
			data.PositionMargin, data.OrderMargin, data.MaintenanceMargin,
			data.AccountEquity, data.AvailableMargin, data.RiskRate, data.PositionCount, data.AssetVersion,
			data.UnrealizedPnl, data.RealizedPnl, data.SourceEventNo, data.SnapshotTime,
			data.Version, data.UpdateTimes, data.Id, current.Version,
		)
	}, idKey, accountKey, oldSourceKey, newSourceKey)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func marginRiskProjectionEqual(a, b *TContractMarginSnapshot) bool {
	return a != nil && b != nil &&
		a.WalletBalance.Equal(b.WalletBalance) &&
		a.AvailableBalance.Equal(b.AvailableBalance) &&
		a.FrozenBalance.Equal(b.FrozenBalance) &&
		a.PositionMargin.Equal(b.PositionMargin) &&
		a.OrderMargin.Equal(b.OrderMargin) &&
		a.MaintenanceMargin.Equal(b.MaintenanceMargin) &&
		a.AccountEquity.Equal(b.AccountEquity) &&
		a.AvailableMargin.Equal(b.AvailableMargin) &&
		a.RiskRate.Equal(b.RiskRate) &&
		a.PositionCount == b.PositionCount &&
		a.AssetVersion == b.AssetVersion &&
		a.UnrealizedPnl.Equal(b.UnrealizedPnl) &&
		a.RealizedPnl.Equal(b.RealizedPnl) &&
		a.SourceEventNo == b.SourceEventNo &&
		a.SnapshotTime == b.SnapshotTime
}

// NewTContractMarginSnapshotModel returns a model for the database table.
func NewTContractMarginSnapshotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractMarginSnapshotModel {
	return &customTContractMarginSnapshotModel{
		defaultTContractMarginSnapshotModel: newTContractMarginSnapshotModel(conn, c, opts...),
	}
}

func (m *defaultTContractMarginSnapshotModel) FindPage(ctx context.Context, tenantId, userId int64, marginAsset string, cursor, limit int64) ([]*TContractMarginSnapshot, int64, error) {
	where, args := []string{"1=1"}, []any{}
	if tenantId > 0 {
		where = append(where, "tenant_id = ?")
		args = append(args, tenantId)
	}
	if userId > 0 {
		where = append(where, "user_id = ?")
		args = append(args, userId)
	}
	if marginAsset != "" {
		where = append(where, "margin_asset = ?")
		args = append(args, marginAsset)
	}
	clause := strings.Join(where, " and ")
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("select count(*) from %s where %s", m.table, clause), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), cursor, limit)
	var rows []*TContractMarginSnapshot
	err := m.QueryRowsNoCacheCtx(ctx, &rows, fmt.Sprintf("select %s from %s where %s and id > ? order by id asc limit ?", tContractMarginSnapshotRows, m.table, clause), queryArgs...)
	return rows, total, err
}
