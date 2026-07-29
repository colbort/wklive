package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractAccountLiquidationItemModel = (*customTContractAccountLiquidationItemModel)(nil)

type (
	// TContractAccountLiquidationItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractAccountLiquidationItemModel.
	TContractAccountLiquidationItemModel interface {
		tContractAccountLiquidationItemModel
		InsertIdempotent(ctx context.Context, data *TContractAccountLiquidationItem) error
		FindByLiquidation(ctx context.Context, tenantID, liquidationID int64, forUpdate bool) ([]*TContractAccountLiquidationItem, error)
		UpdateStatus(ctx context.Context, id, status, updateTimes int64) error
		UpdateADL(ctx context.Context, data *TContractAccountLiquidationItem) error
	}

	customTContractAccountLiquidationItemModel struct {
		*defaultTContractAccountLiquidationItemModel
		conn sqlx.SqlConn
	}
)

// NewTContractAccountLiquidationItemModel returns a model for the database table.
func NewTContractAccountLiquidationItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractAccountLiquidationItemModel {
	return &customTContractAccountLiquidationItemModel{
		defaultTContractAccountLiquidationItemModel: newTContractAccountLiquidationItemModel(conn, c, opts...),
		conn: conn,
	}
}

func (m *customTContractAccountLiquidationItemModel) InsertIdempotent(
	ctx context.Context, data *TContractAccountLiquidationItem,
) error {
	_, err := m.conn.ExecCtx(ctx, `INSERT INTO t_contract_account_liquidation_item
(tenant_id,account_liquidation_id,liquidation_no,position_id,position_version,symbol_id,position_side,trigger_qty,trigger_mark_price,trigger_snapshot_id,position_margin,maintenance_margin,realized_pnl,liquidation_fee,deficit_amount,bankruptcy_price,adl_relief_amount,adl_qty,status,create_times,update_times)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE id=id`,
		data.TenantId, data.AccountLiquidationId, data.LiquidationNo,
		data.PositionId, data.PositionVersion, data.SymbolId, data.PositionSide,
		data.TriggerQty, data.TriggerMarkPrice, data.TriggerSnapshotId,
		data.PositionMargin, data.MaintenanceMargin, data.RealizedPnl,
		data.LiquidationFee, data.DeficitAmount, data.BankruptcyPrice,
		data.AdlReliefAmount, data.AdlQty, data.Status,
		data.CreateTimes, data.UpdateTimes,
	)
	return err
}

func (m *customTContractAccountLiquidationItemModel) FindByLiquidation(
	ctx context.Context, tenantID, liquidationID int64, forUpdate bool,
) ([]*TContractAccountLiquidationItem, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id=? AND account_liquidation_id=? ORDER BY position_id",
		tContractAccountLiquidationItemRows, m.table,
	)
	if forUpdate {
		query += " FOR UPDATE"
	}
	var rows []*TContractAccountLiquidationItem
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, tenantID, liquidationID); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customTContractAccountLiquidationItemModel) UpdateStatus(
	ctx context.Context, id, status, updateTimes int64,
) error {
	_, err := m.conn.ExecCtx(
		ctx, fmt.Sprintf("UPDATE %s SET status=?,update_times=? WHERE id=?", m.table),
		status, updateTimes, id,
	)
	return err
}

func (m *customTContractAccountLiquidationItemModel) UpdateADL(
	ctx context.Context, data *TContractAccountLiquidationItem,
) error {
	result, err := m.conn.ExecCtx(ctx, fmt.Sprintf(`UPDATE %s
SET deficit_amount=?,bankruptcy_price=?,adl_relief_amount=?,adl_qty=?,update_times=?
WHERE id=?`, m.table),
		data.DeficitAmount, data.BankruptcyPrice, data.AdlReliefAmount,
		data.AdlQty, data.UpdateTimes, data.Id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf(
			"account liquidation item %d update affected %d rows",
			data.Id, affected,
		)
	}
	return nil
}
