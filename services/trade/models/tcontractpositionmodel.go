package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
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

	// TContractPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionModel.
	TContractPositionModel interface {
		tContractPositionModel
		FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error)
		FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error)
		FindActiveListForUpdate(ctx context.Context, tenantID, symbolID int64) ([]*TContractPosition, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TContractPosition, error)
		FindOneForUpdateByTenantUserSymbolSideMode(ctx context.Context, tenantID, userID, symbolID, positionSide, marginMode int64) (*TContractPosition, error)
		ReserveCloseQty(ctx context.Context, id, version int64, qty decimal.Decimal, updateTimes int64) error
		ReleaseCloseQty(ctx context.Context, id int64, qty decimal.Decimal, updateTimes int64) error
		UpdateMarkRiskCAS(ctx context.Context, data *TContractPosition, expectedVersion, updateTimes int64) (bool, error)
	}

	customTContractPositionModel struct {
		*defaultTContractPositionModel
	}
)

// NewTContractPositionModel returns a model for the database table.
func NewTContractPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionModel {
	return &customTContractPositionModel{
		defaultTContractPositionModel: newTContractPositionModel(conn, c, opts...),
	}
}

// UpdateMarkRiskCAS updates only mark-derived fields. It deliberately avoids
// writing quantity, margin or realized PnL read by the scanner, and the version
// predicate prevents a stale mark refresh from overwriting a concurrent Fill.
func (m *defaultTContractPositionModel) UpdateMarkRiskCAS(ctx context.Context, data *TContractPosition, expectedVersion, updateTimes int64) (bool, error) {
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

func (m *defaultTContractPositionModel) FindActiveListForUpdate(ctx context.Context, tenantID, symbolID int64) ([]*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND symbol_id = ? AND status = 1 AND qty > 0 ORDER BY id FOR UPDATE", tContractPositionRows, m.table)
	var positions []*TContractPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &positions, query, tenantID, symbolID); err != nil {
		return nil, err
	}
	return positions, nil
}

func (m *defaultTContractPositionModel) FindOneForUpdate(ctx context.Context, id int64) (*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tContractPositionRows, m.table)
	var position TContractPosition
	if err := m.QueryRowNoCacheCtx(ctx, &position, query, id); err != nil {
		return nil, err
	}
	return &position, nil
}

func (m *defaultTContractPositionModel) FindOneForUpdateByTenantUserSymbolSideMode(ctx context.Context, tenantID, userID, symbolID, positionSide, marginMode int64) (*TContractPosition, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND user_id = ? AND symbol_id = ? AND position_side = ? AND margin_mode = ? LIMIT 1 FOR UPDATE", tContractPositionRows, m.table)
	var position TContractPosition
	if err := m.QueryRowNoCacheCtx(ctx, &position, query, tenantID, userID, symbolID, positionSide, marginMode); err != nil {
		return nil, err
	}
	return &position, nil
}

func (m *defaultTContractPositionModel) ReleaseCloseQty(ctx context.Context, id int64, qty decimal.Decimal, updateTimes int64) error {
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

func (m *defaultTContractPositionModel) ReserveCloseQty(ctx context.Context, id, version int64, qty decimal.Decimal, updateTimes int64) error {
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

func (m *defaultTContractPositionModel) FindPage(ctx context.Context, filter ContractPositionPageFilter, cursor int64, limit int64) ([]*TContractPosition, int64, error) {
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

func (m *defaultTContractPositionModel) FindList(ctx context.Context, filter ContractPositionPageFilter) ([]*TContractPosition, error) {
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
