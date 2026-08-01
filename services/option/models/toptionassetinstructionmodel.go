package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (m *customTOptionAssetInstructionModel) Insert(
	ctx context.Context,
	data *TOptionAssetInstruction,
) (sql.Result, error) {
	if data == nil || strings.TrimSpace(data.Coin) == "" {
		return nil, fmt.Errorf("option asset instruction coin is required")
	}
	data.Coin = strings.TrimSpace(data.Coin)
	return m.defaultTOptionAssetInstructionModel.Insert(ctx, data)
}

var _ TOptionAssetInstructionModel = (*customTOptionAssetInstructionModel)(nil)

type (
	OptionAssetInstructionPageFilter struct {
		TenantId             int64
		UserId               int64
		BizNo                string
		Status               int64
		ReconciliationStatus int64
	}

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
		FindByDeliveryUnit(ctx context.Context, tenantId, deliveryUnitId int64) ([]*TOptionAssetInstruction, error)
		FindByComboOrderID(ctx context.Context, tenantId, comboOrderId, limit int64) ([]*TOptionAssetInstruction, int64, error)
		ResetFailedByDeliveryUnit(ctx context.Context, tenantId, deliveryUnitId, now int64) (int64, error)
		FindPage(ctx context.Context, filter OptionAssetInstructionPageFilter, cursor, limit int64) ([]*TOptionAssetInstruction, int64, error)
		RecoverStale(ctx context.Context, tenantId, staleBefore, now int64) (int64, error)
	}

	customTOptionAssetInstructionModel struct {
		*defaultTOptionAssetInstructionModel
	}
)

func (m *defaultTOptionAssetInstructionModel) RecoverStale(
	ctx context.Context, tenantId, staleBefore, now int64,
) (int64, error) {
	query := `UPDATE t_option_asset_instruction
SET status=?,next_retry_at=?,last_error_msg='STALE_PROCESSING_RECOVERED',update_times=?
WHERE status=? AND update_times < ?`
	args := []any{
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		now, now,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING),
		staleBefore,
	}
	if tenantId > 0 {
		query += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	result, err := m.ExecNoCacheCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *defaultTOptionAssetInstructionModel) FindPage(
	ctx context.Context, filter OptionAssetInstructionPageFilter, cursor, limit int64,
) ([]*TOptionAssetInstruction, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.LikeString("biz_no", filter.BizNo)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("reconciliation_status", filter.ReconciliationStatus)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionAssetInstructionRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionAssetInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *customTOptionAssetInstructionModel) FindByDeliveryUnit(
	ctx context.Context, tenantId, deliveryUnitId int64,
) ([]*TOptionAssetInstruction, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND delivery_unit_id = ? ORDER BY step_no, id`,
		tOptionAssetInstructionRows, m.table)
	var list []*TOptionAssetInstruction
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, deliveryUnitId)
	return list, err
}

func (m *defaultTOptionAssetInstructionModel) FindByComboOrderID(
	ctx context.Context, tenantId, comboOrderId, limit int64,
) ([]*TOptionAssetInstruction, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where := `instruction.tenant_id=? AND EXISTS (
  SELECT 1 FROM t_option_order AS child
  WHERE child.tenant_id=instruction.tenant_id AND child.combo_order_id=?
    AND child.id=instruction.order_id
)`
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s AS instruction WHERE %s", m.table, where),
		tenantId, comboOrderId,
	); err != nil {
		return nil, 0, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s AS instruction
WHERE %s
ORDER BY instruction.order_id,instruction.step_no,instruction.id
LIMIT ?`, tOptionAssetInstructionRows, m.table, where)
	var list []*TOptionAssetInstruction
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId, limit); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTOptionAssetInstructionModel) ResetFailedByDeliveryUnit(
	ctx context.Context, tenantId, deliveryUnitId, now int64,
) (int64, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_option_asset_instruction
SET status = ?, retry_count = 0, next_retry_at = ?, last_error_msg = '',
    reconciliation_status = ?, update_times = ?
WHERE tenant_id = ? AND delivery_unit_id = ? AND status IN (?, ?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING), now,
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING), now,
		tenantId, deliveryUnitId,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

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
      AND COALESCE(NULLIF(biz_previous.execution_group, ''), biz_previous.biz_no)
          = COALESCE(NULLIF(current.execution_group, ''), current.biz_no)
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
  )
  AND (
    current.delivery_unit_id = 0
    OR EXISTS (
      SELECT 1 FROM t_option_physical_delivery_unit unit
      WHERE unit.tenant_id = current.tenant_id
        AND unit.id = current.delivery_unit_id
        AND unit.status IN (1, 2, 3)
        AND (
          unit.cure_deadline > ?
          OR (unit.status = 2 AND unit.manual_retry_count > 0)
        )
    )
  )`, tOptionAssetInstructionRows, m.table, m.table, m.table)
	args := []any{
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		now, cursor,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		now,
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
