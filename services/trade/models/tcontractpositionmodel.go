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

var _ TContractPositionModel = (*customTContractPositionModel)(nil)

type (
	ContractPositionPageFilter struct {
		TenantId     int64
		UserId       int64
		SymbolId     int64
		ContractType int64
		PositionSide int64
	}

	CrossMarginOpeningAggregate struct {
		PositionMargin    decimal.Decimal `db:"position_margin"`
		MaintenanceMargin decimal.Decimal `db:"maintenance_margin"`
		UnrealizedPnl     decimal.Decimal `db:"unrealized_pnl"`
		StaleMarkCount    int64           `db:"stale_mark_count"`
	}

	// TContractPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionModel.
	TContractPositionModel interface {
		tContractPositionModel
		FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error)
		FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error)
		FindActiveListForUpdate(ctx context.Context, tenantID, symbolID int64) ([]*TContractPosition, error)
		FindCrossRiskUnitForUpdate(ctx context.Context, tenantID, userID int64, marginAsset string) ([]*TContractPosition, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TContractPosition, error)
		FindOneForUpdateByTenantUserSymbolSideMode(ctx context.Context, tenantID, userID, symbolID, positionSide, marginMode int64) (*TContractPosition, error)
		CountActiveIncompatibleMode(ctx context.Context, tenantID, userID, symbolID, marginMode, positionMode int64) (int64, error)
		CountOpenRiskUnit(ctx context.Context, tenantID, userID, symbolID int64) (int64, error)
		FindCrossMarginOpeningAggregate(ctx context.Context, tenantID, userID int64, marginAsset string, minMarkTime int64) (*CrossMarginOpeningAggregate, error)
		ReserveCloseQty(ctx context.Context, id, version int64, qty decimal.Decimal, updateTimes int64) error
		ReleaseCloseQty(ctx context.Context, id int64, qty decimal.Decimal, updateTimes int64) error
		UpdateMarkRiskCAS(ctx context.Context, data *TContractPosition, expectedVersion, updateTimes int64) (bool, error)
	}

	customTContractPositionModel struct {
		*defaultTContractPositionModel
	}
)

func (m *customTContractPositionModel) CountActiveIncompatibleMode(
	ctx context.Context, tenantID, userID, symbolID, marginMode, positionMode int64,
) (int64, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_contract_position
WHERE tenant_id=? AND user_id=? AND symbol_id=? AND qty>0
  AND status IN (1,2,3,4,6)
  AND (margin_mode<>? OR position_mode<>?)`,
		tenantID, userID, symbolID, marginMode, positionMode)
	return count, err
}

func (m *customTContractPositionModel) CountOpenRiskUnit(
	ctx context.Context, tenantID, userID, symbolID int64,
) (int64, error) {
	query := `SELECT COUNT(1) FROM t_contract_position
WHERE tenant_id=? AND user_id=? AND status IN (1,2,3,4,6) AND qty>0`
	args := []any{tenantID, userID}
	if symbolID > 0 {
		query += " AND symbol_id=?"
		args = append(args, symbolID)
	}
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customTContractPositionModel) FindCrossMarginOpeningAggregate(
	ctx context.Context, tenantID, userID int64, marginAsset string, minMarkTime int64,
) (*CrossMarginOpeningAggregate, error) {
	var aggregate CrossMarginOpeningAggregate
	err := m.QueryRowNoCacheCtx(ctx, &aggregate, `SELECT
  COALESCE(SUM(p.position_margin),0) AS position_margin,
  COALESCE(SUM(p.maintenance_margin),0) AS maintenance_margin,
  COALESCE(SUM(p.unrealized_pnl),0) AS unrealized_pnl,
  COALESCE(SUM(CASE
    WHEN p.mark_snapshot_id='' OR NOT EXISTS (
      SELECT 1
      FROM t_trade_market_snapshot s
      WHERE s.tenant_id=p.tenant_id
        AND s.symbol_id=p.symbol_id
        AND s.snapshot_id=p.mark_snapshot_id
        AND s.snapshot_kind='MARK'
        AND s.confirmed=1
        AND s.source_timestamp>=?
    ) THEN 1 ELSE 0 END),0) AS stale_mark_count
FROM t_contract_position p
WHERE p.tenant_id=? AND p.user_id=? AND p.margin_asset=?
  AND p.margin_mode=1 AND p.status=1 AND p.qty>0`,
		minMarkTime, tenantID, userID, marginAsset)
	if err != nil {
		return nil, err
	}
	return &aggregate, nil
}

// NewTContractPositionModel returns a model for the database table.
func NewTContractPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionModel {
	return &customTContractPositionModel{
		defaultTContractPositionModel: newTContractPositionModel(conn, c, opts...),
	}
}

// UpdateMarkRiskCAS updates only mark-derived fields. It deliberately avoids
// writing quantity, margin or realized PnL read by the scanner, and the version
// predicate prevents a stale mark refresh from overwriting a concurrent Fill.
func (m *customTContractPositionModel) UpdateMarkRiskCAS(ctx context.Context, data *TContractPosition, expectedVersion, updateTimes int64) (bool, error) {
	current, err := m.FindOne(ctx, data.Id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTContractPositionIdPrefix, current.Id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v:%v:%v", cacheTContractPositionTenantIdUserIdSymbolIdPositionSideMarginModePrefix, current.TenantId, current.UserId, current.SymbolId, current.PositionSide, current.MarginMode)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf(`UPDATE %s SET
			mark_price=?,mark_snapshot_id=?,maintenance_margin=?,unrealized_pnl=?,
			liquidation_price=?,bankruptcy_price=?,risk_rate=?,adl_rank=?,
			version=version+1,update_times=?
			WHERE id=? AND version=? AND status=1 AND qty>0`, m.table)
		return conn.ExecCtx(ctx, query,
			data.MarkPrice, data.MarkSnapshotId, data.MaintenanceMargin, data.UnrealizedPnl,
			data.LiquidationPrice, data.BankruptcyPrice, data.RiskRate, data.AdlRank,
			updateTimes, data.Id, expectedVersion,
		)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTContractPositionModel) FindActiveListForUpdate(ctx context.Context, tenantID, symbolID int64) ([]*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND symbol_id = ? AND status = 1 AND qty > 0 ORDER BY id FOR UPDATE", tContractPositionRows, m.table)
	var positions []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &positions, query, tenantID, symbolID); err != nil {
		return nil, err
	}
	return positions, nil
}

func (m *customTContractPositionModel) FindCrossRiskUnitForUpdate(ctx context.Context, tenantID, userID int64, marginAsset string) ([]*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND user_id=? AND margin_asset=? AND margin_mode=1 AND status=1 AND qty>0 ORDER BY id FOR UPDATE", tContractPositionRows, m.table)
	var positions []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &positions, query, tenantID, userID, marginAsset); err != nil {
		return nil, err
	}
	return positions, nil
}

func (m *customTContractPositionModel) FindOneForUpdate(ctx context.Context, id int64) (*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tContractPositionRows, m.table)
	var position TContractPosition
	if err := m.QueryRowNoCacheCtx(ctx, &position, query, id); err != nil {
		return nil, err
	}
	return &position, nil
}

func (m *customTContractPositionModel) FindOneForUpdateByTenantUserSymbolSideMode(ctx context.Context, tenantID, userID, symbolID, positionSide, marginMode int64) (*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND user_id = ? AND symbol_id = ? AND position_side = ? AND margin_mode = ? LIMIT 1 FOR UPDATE", tContractPositionRows, m.table)
	var position TContractPosition
	if err := m.QueryRowNoCacheCtx(ctx, &position, query, tenantID, userID, symbolID, positionSide, marginMode); err != nil {
		return nil, err
	}
	return &position, nil
}

func (m *customTContractPositionModel) ReleaseCloseQty(ctx context.Context, id int64, qty decimal.Decimal, updateTimes int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	idKey := fmt.Sprintf("%s%v", cacheTContractPositionIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v:%v:%v", cacheTContractPositionTenantIdUserIdSymbolIdPositionSideMarginModePrefix, data.TenantId, data.UserId, data.SymbolId, data.PositionSide, data.MarginMode)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET avail_qty = avail_qty + ?, frozen_qty = frozen_qty - ?, version = version + 1, update_times = ? WHERE id = ? AND frozen_qty >= ?", m.table)
		return conn.ExecCtx(ctx, query, qty, qty, updateTimes, id, qty)
	}, idKey, uniqueKey)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("reserved position quantity is insufficient")
	}
	return nil
}

func (m *customTContractPositionModel) ReserveCloseQty(ctx context.Context, id, version int64, qty decimal.Decimal, updateTimes int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	idKey := fmt.Sprintf("%s%v", cacheTContractPositionIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v:%v:%v", cacheTContractPositionTenantIdUserIdSymbolIdPositionSideMarginModePrefix, data.TenantId, data.UserId, data.SymbolId, data.PositionSide, data.MarginMode)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET avail_qty = avail_qty - ?, frozen_qty = frozen_qty + ?, version = version + 1, update_times = ? WHERE id = ? AND version = ? AND status = 1 AND avail_qty >= ?", m.table)
		return conn.ExecCtx(ctx, query, qty, qty, updateTimes, id, version, qty)
	}, idKey, uniqueKey)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("position changed or available quantity is insufficient")
	}
	return nil
}

func (m *customTContractPositionModel) FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("contract_type", filter.ContractType)
	builder.EqInt64("position_side", filter.PositionSide)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractPositionRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *customTContractPositionModel) FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("contract_type", filter.ContractType)
	builder.EqInt64("position_side", filter.PositionSide)

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC", tContractPositionRows, m.table, builder.Where())
	var list []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
