package models

import (
	"context"
	"fmt"

	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionAssetInstructionModel = (*customTOptionAssetInstructionModel)(nil)

type (
	// TOptionAssetInstructionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionAssetInstructionModel.
	TOptionAssetInstructionModel interface {
		tOptionAssetInstructionModel
		FindRunnable(ctx context.Context, tenantId, now, cursor, limit int64) ([]*TOptionAssetInstruction, error)
		FindByBizNo(ctx context.Context, tenantId int64, bizNo string) ([]*TOptionAssetInstruction, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionAssetInstruction, error)
		Claim(ctx context.Context, id, now int64) (bool, error)
		ResetForManualRetry(ctx context.Context, id, now int64) (bool, error)
		ResetFailedByBizNo(ctx context.Context, tenantId int64, bizNo string, now int64) (int64, error)
		HasIncompleteMarginForContract(ctx context.Context, tenantId, contractId int64) (bool, error)
		FindByLiquidation(ctx context.Context, tenantId, liquidationId int64) ([]*TOptionAssetInstruction, error)
	}

	customTOptionAssetInstructionModel struct {
		*defaultTOptionAssetInstructionModel
	}
)

func (m *customTOptionAssetInstructionModel) FindByLiquidation(ctx context.Context, tenantId, liquidationId int64) ([]*TOptionAssetInstruction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND liquidation_id = ? ORDER BY step_no, id",
		tOptionAssetInstructionRows, m.table)
	var list []*TOptionAssetInstruction
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, liquidationId)
	return list, err
}

func (m *customTOptionAssetInstructionModel) HasIncompleteMarginForContract(ctx context.Context, tenantId, contractId int64) (bool, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_option_asset_instruction i
JOIN t_option_margin_lot l ON l.tenant_id = i.tenant_id AND l.id = i.margin_lot_id
WHERE i.tenant_id = ? AND l.contract_id = ? AND i.status <> ?`,
		tenantId, contractId, int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
	)
	return count > 0, err
}

func (m *defaultTOptionAssetInstructionModel) ResetForManualRetry(ctx context.Context, id, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_asset_instruction
SET status = ?, retry_count = 0, next_retry_at = ?, last_error_msg = '',
    reconciliation_status = ?, update_times = ?
WHERE id = ? AND status IN (?, ?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING), now,
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING), now, id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *defaultTOptionAssetInstructionModel) ResetFailedByBizNo(
	ctx context.Context,
	tenantId int64,
	bizNo string,
	now int64,
) (int64, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_asset_instruction
SET status = ?, retry_count = 0, next_retry_at = ?, last_error_msg = '',
    reconciliation_status = ?, update_times = ?
WHERE tenant_id = ? AND biz_no = ? AND status IN (?, ?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING), now,
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING), now,
		tenantId, bizNo,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// NewTOptionAssetInstructionModel returns a model for the database table.
func NewTOptionAssetInstructionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionAssetInstructionModel {
	return &customTOptionAssetInstructionModel{
		defaultTOptionAssetInstructionModel: newTOptionAssetInstructionModel(conn, c, opts...),
	}
}

func (m *customTOptionAssetInstructionModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionAssetInstruction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionAssetInstructionRows, m.table)
	var item TOptionAssetInstruction
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionAssetInstructionModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_asset_instruction
SET status=?,update_times=?
WHERE id=? AND status IN (?,?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING), now, id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED))
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTOptionAssetInstructionModel) FindByBizNo(ctx context.Context, tenantId int64, bizNo string) ([]*TOptionAssetInstruction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND biz_no = ? ORDER BY step_no ASC, id ASC", tOptionAssetInstructionRows, m.table)
	var items []*TOptionAssetInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, bizNo); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customTOptionAssetInstructionModel) FindRunnable(ctx context.Context, tenantId, now, cursor, limit int64) ([]*TOptionAssetInstruction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := fmt.Sprintf(`SELECT %s FROM %s AS current
WHERE current.status IN (?, ?) AND current.next_retry_at <= ? AND current.id > ?
  AND NOT EXISTS (
    SELECT 1 FROM %s AS biz_previous
    WHERE biz_previous.tenant_id = current.tenant_id
      AND biz_previous.biz_no = current.biz_no
      AND biz_previous.step_no < current.step_no
      AND biz_previous.status <> ?
  )
  AND NOT EXISTS (
    SELECT 1 FROM %s AS previous
    WHERE previous.tenant_id = current.tenant_id
      AND previous.order_id = current.order_id
      AND current.order_id > 0
      AND previous.id < current.id
      AND previous.action IN (2, 3)
      AND current.action IN (2, 3)
      AND previous.status <> ?
  )`, tOptionAssetInstructionRows, m.table, m.table, m.table)
	args := []any{
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		now, cursor,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
	}
	if tenantId > 0 {
		query += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	query += " ORDER BY id ASC LIMIT ?"
	args = append(args, limit)
	var items []*TOptionAssetInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}
