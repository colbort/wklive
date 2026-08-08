package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractReconciliationCursorModel = (*customTContractReconciliationCursorModel)(nil)

type (
	TContractReconciliationCursorModel interface {
		tContractReconciliationCursorModel
		FindContractOrderFillAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		FindContractFillPositionAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		FindContractReservationAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		FindContractPositionMarginAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		FindContractLiquidationAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		FindCrossAccountLiquidationAudits(ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int) error
		LoadReconciliationCursor(ctx context.Context, tenantID int64, checkType string, now int64) (int64, error)
		AdvanceReconciliationCursor(ctx context.Context, tenantID int64, checkType string, cursor, now int64) error
		CompleteReconciliationCycle(ctx context.Context, tenantID int64, checkType string, now int64) error
	}

	customTContractReconciliationCursorModel struct {
		*defaultTContractReconciliationCursorModel
	}
)

func NewTContractReconciliationCursorModel(
	conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option,
) TContractReconciliationCursorModel {
	return &customTContractReconciliationCursorModel{
		defaultTContractReconciliationCursorModel: newTContractReconciliationCursorModel(conn, c, opts...),
	}
}

func (m *customTContractReconciliationCursorModel) LoadReconciliationCursor(
	ctx context.Context, tenantID int64, checkType string, now int64,
) (int64, error) {
	_, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_contract_reconciliation_cursor
(tenant_id,check_type,last_scanned_id,last_cycle_completed_at,create_times,update_times)
VALUES(?,?,0,0,?,?)
ON DUPLICATE KEY UPDATE update_times=update_times`, tenantID, checkType, now, now)
	if err != nil {
		return 0, err
	}
	var cursor int64
	err = m.QueryRowNoCacheCtx(ctx, &cursor,
		"SELECT last_scanned_id FROM t_contract_reconciliation_cursor WHERE tenant_id=? AND check_type=? LIMIT 1",
		tenantID, checkType)
	return cursor, err
}

func (m *customTContractReconciliationCursorModel) AdvanceReconciliationCursor(
	ctx context.Context, tenantID int64, checkType string, cursor, now int64,
) error {
	_, err := m.ExecNoCacheCtx(ctx,
		"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=?,update_times=? WHERE tenant_id=? AND check_type=?",
		cursor, now, tenantID, checkType)
	return err
}

func (m *customTContractReconciliationCursorModel) CompleteReconciliationCycle(
	ctx context.Context, tenantID int64, checkType string, now int64,
) error {
	_, err := m.ExecNoCacheCtx(ctx,
		"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=0,last_cycle_completed_at=?,update_times=? WHERE tenant_id=? AND check_type=?",
		now, now, tenantID, checkType)
	return err
}

func reconciliationTenantClause(alias string, tenantID int64, args []any) (string, []any) {
	if tenantID <= 0 {
		return "", args
	}
	return " AND " + alias + ".tenant_id=?", append(args, tenantID)
}

func (m *customTContractReconciliationCursorModel) FindContractOrderFillAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	tenantClause, args := reconciliationTenantClause("o", tenantID, []any{int64(2), cursor, cutoff})
	args = append(args, limit)
	query := `
SELECT
  o.tenant_id,o.id AS order_id,o.order_no,o.status,o.contract_value_type,s.price_scale,
  o.qty AS order_qty,o.filled_qty AS order_filled_qty,o.filled_amount AS order_filled_amount,
  o.canceled_qty AS order_canceled_qty,o.avg_price AS order_avg_price,o.fee AS order_fee,
  COUNT(f.id) AS fill_count,
  COALESCE(SUM(f.qty),0) AS fill_qty,
  COALESCE(SUM(f.amount),0) AS fill_amount,
  COALESCE(SUM(f.fee),0) AS fill_fee,
  COALESCE(CASE
    WHEN SUM(f.qty) IS NULL OR SUM(f.qty)=0 THEN 0
    WHEN o.contract_value_type=2 THEN SUM(f.qty)/SUM(f.qty/f.price)
    ELSE SUM(f.price*f.qty)/SUM(f.qty)
  END,0) AS fill_avg_price
FROM t_trade_order o
JOIN t_trade_symbol s ON s.tenant_id=o.tenant_id AND s.id=o.symbol_id
LEFT JOIN t_trade_fill f ON f.tenant_id=o.tenant_id AND f.order_id=o.id
WHERE o.product_type=? AND o.id>? AND o.update_times<=?` + tenantClause + `
GROUP BY o.tenant_id,o.id,o.order_no,o.status,o.contract_value_type,s.price_scale,o.qty,o.filled_qty,
         o.filled_amount,o.canceled_qty,o.avg_price,o.fee
ORDER BY o.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTContractReconciliationCursorModel) FindContractFillPositionAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	tenantClause, args := reconciliationTenantClause("f", tenantID, []any{int64(2), cursor, cutoff})
	args = append(args, limit)
	query := `
SELECT
  f.id,f.tenant_id,f.fill_no,f.order_id,f.user_id,f.symbol_id,f.position_side,
  f.qty AS fill_qty,f.price AS fill_price,f.fee AS fill_fee,f.fee_asset,
  f.create_times AS fill_create_times,
  COUNT(h.id) AS history_count,
  COALESCE(SUM(ABS(h.after_qty-h.before_qty)),0) AS projected_qty,
  COALESCE(SUM(h.fee_delta),0) AS projected_fee,
  COALESCE(SUM(CASE WHEN h.tenant_id<>f.tenant_id OR h.user_id<>f.user_id OR
                              h.symbol_id<>f.symbol_id OR h.ref_order_id<>f.order_id OR
                              h.ref_fill_id<>f.id THEN 1 ELSE 0 END),0) AS identity_mismatch,
  COALESCE(SUM(CASE WHEN h.mark_price<>f.price THEN 1 ELSE 0 END),0) AS price_mismatch,
  COALESCE(SUM(CASE WHEN h.business_time<>f.create_times THEN 1 ELSE 0 END),0) AS time_mismatch,
  COALESCE(SUM(CASE WHEN h.after_version<>h.before_version+1 THEN 1 ELSE 0 END),0) AS version_mismatch,
  COALESCE(SUM(CASE WHEN f.fee<>0 AND h.fee_asset<>f.fee_asset THEN 1 ELSE 0 END),0) AS fee_asset_mismatch
FROM t_trade_fill f
LEFT JOIN t_contract_position_history h
  ON h.tenant_id=f.tenant_id AND h.ref_fill_id=f.id
WHERE f.product_type=? AND f.id>? AND f.create_times<=?` + tenantClause + `
GROUP BY f.id,f.tenant_id,f.fill_no,f.order_id,f.user_id,f.symbol_id,f.position_side,
         f.qty,f.price,f.fee,f.fee_asset,f.create_times
ORDER BY f.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTContractReconciliationCursorModel) FindContractReservationAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	tenantClause, args := reconciliationTenantClause("r", tenantID, []any{int64(2), cursor, cutoff})
	args = append(args, limit)
	query := `
SELECT r.id,r.tenant_id,r.order_id,o.order_no,o.user_id,r.reservation_no,r.asset,
       r.reserved_amount,r.consumed_amount,r.released_amount,r.status
FROM t_trade_asset_reservation r
JOIN t_trade_order o ON o.tenant_id=r.tenant_id AND o.id=r.order_id
WHERE o.product_type=? AND r.id>? AND r.update_times<=?` + tenantClause + `
ORDER BY r.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTContractReconciliationCursorModel) FindContractPositionMarginAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	args := []any{6, 3, 7, 3, 3, 5, cursor, cutoff}
	tenantClause, args := reconciliationTenantClause("p", tenantID, args)
	args = append(args, limit)
	query := `
SELECT
  p.id,p.tenant_id,p.user_id,p.symbol_id,p.margin_asset,p.position_margin,p.isolated_margin,
  COALESCE(SUM(CASE WHEN i.action=? AND i.status=? AND i.reconciled_at>0 THEN i.amount ELSE 0 END),0) AS margin_consumed,
  COALESCE(SUM(CASE WHEN i.action=? AND i.status=? AND i.reconciled_at>0 THEN i.amount ELSE 0 END),0) AS margin_released,
  COALESCE(SUM(CASE WHEN i.id IS NOT NULL AND (i.status<>? OR i.reconciled_at=0) AND i.action IN (6,7) THEN 1 ELSE 0 END),0) AS unfinished_count,
  COALESCE((SELECT COUNT(1) FROM t_contract_position_history h
            WHERE h.tenant_id=p.tenant_id AND h.position_id=p.id AND h.action_type=?),0) AS liquidation_count
FROM t_contract_position p
LEFT JOIN t_trade_settlement_instruction i
  ON i.tenant_id=p.tenant_id AND i.position_id=p.id AND i.action IN (6,7)
WHERE p.id>? AND p.update_times<=?` + tenantClause + `
GROUP BY p.id,p.tenant_id,p.user_id,p.symbol_id,p.margin_asset,p.position_margin,p.isolated_margin
ORDER BY p.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTContractReconciliationCursorModel) FindContractLiquidationAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	args := []any{5, 3, cursor, cutoff}
	tenantClause, args := reconciliationTenantClause("q", tenantID, args)
	args = append(args, limit)
	query := `
SELECT
  q.id,q.tenant_id,q.liquidation_no,q.position_id,q.status,q.trigger_qty,q.liquidated_qty,
  q.insurance_fund_amount,q.adl_qty,q.completed_at,
  COALESCE(p.qty,0) AS position_qty,COALESCE(p.position_margin,0) AS position_margin,
  COALESCE(p.isolated_margin,0) AS isolated_margin,COALESCE(p.status,0) AS position_status,
  COALESCE(p.margin_asset,'') AS margin_asset,
  COALESCE((SELECT COUNT(1) FROM t_contract_position_history h
            WHERE h.tenant_id=q.tenant_id AND h.position_id=q.position_id
              AND h.action_key=q.liquidation_no AND h.action_type=?),0) AS liquidation_history,
  COALESCE((SELECT COUNT(1) FROM t_trade_event_outbox e
            WHERE e.tenant_id=q.tenant_id AND e.event_no=CONCAT(q.liquidation_no,'-COMPLETED')),0) AS completion_event,
  COALESCE((SELECT COUNT(1) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_execution_count,
  COALESCE((SELECT SUM(a.status=3) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_completed_count,
  COALESCE((SELECT SUM(a.status IN (4,5)) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_failed_count,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.qty ELSE 0 END) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_execution_qty,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.relief_amount ELSE 0 END) FROM t_contract_adl_execution a
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id),0) AS adl_relief_amount,
  COALESCE((SELECT COUNT(1) FROM t_contract_adl_execution a
            LEFT JOIN t_trade_settlement_instruction i
              ON i.tenant_id=a.tenant_id AND i.instruction_no=a.execution_no
            WHERE a.tenant_id=q.tenant_id AND a.liquidation_id=q.id AND a.asset_credit>0
              AND (i.id IS NULL OR i.status<>? OR i.reconciled_at=0)),0) AS adl_unreconciled_assets
FROM t_contract_liquidation q
LEFT JOIN t_contract_position p ON p.tenant_id=q.tenant_id AND p.id=q.position_id
WHERE q.id>? AND q.update_times<=?` + tenantClause + `
ORDER BY q.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTContractReconciliationCursorModel) FindCrossAccountLiquidationAudits(
	ctx context.Context, dest any, tenantID, cursor, cutoff int64, limit int,
) error {
	args := []any{2, 5, 5, 3, 3, 3, cursor, cutoff}
	tenantClause, args := reconciliationTenantClause("q", tenantID, args)
	args = append(args, limit)
	query := `
SELECT
  q.id,q.tenant_id,q.liquidation_no,q.margin_asset,q.status,q.position_count,q.started_at,q.completed_at,
  q.gross_settlement,q.position_margin,q.liquidation_fee,q.user_credit,q.user_debit,
  q.deficit_amount,q.insurance_fund_amount,q.adl_relief_amount,q.adl_qty,
  COALESCE((SELECT COUNT(1) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_count,
  COALESCE((SELECT SUM(i.status=?) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS closed_item_count,
  COALESCE((SELECT SUM(p.status=? AND p.qty=0 AND p.position_margin=0 AND
                             p.isolated_margin=0 AND p.maintenance_margin=0)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_position p ON p.tenant_id=i.tenant_id AND p.id=i.position_id
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS closed_position_count,
  COALESCE((SELECT COUNT(1)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_position_history h
              ON h.tenant_id=i.tenant_id AND h.position_id=i.position_id
             AND h.action_key=CONCAT(q.liquidation_no,'-',i.position_id)
             AND h.action_type=?
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS history_count,
  COALESCE((SELECT COUNT(1) FROM t_trade_event_outbox e
            WHERE e.tenant_id=q.tenant_id
              AND e.event_no=CONCAT(q.liquidation_no,'-COMPLETED')),0) AS completion_event,
  COALESCE((SELECT SUM(i.position_margin) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_position_margin,
  COALESCE((SELECT SUM(i.realized_pnl) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_realized_pnl,
  COALESCE((SELECT SUM(i.liquidation_fee) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_fee,
  COALESCE((SELECT SUM(i.deficit_amount) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_deficit,
  COALESCE((SELECT SUM(i.adl_relief_amount) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_adl_relief,
  COALESCE((SELECT SUM(i.adl_qty) FROM t_contract_account_liquidation_item i
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS item_adl_qty,
  COALESCE((SELECT COUNT(1)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_adl_execution a
              ON a.tenant_id=i.tenant_id AND a.liquidation_id=-i.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',i.id)
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS adl_execution_count,
  COALESCE((SELECT SUM(a.status=3)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_adl_execution a
              ON a.tenant_id=i.tenant_id AND a.liquidation_id=-i.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',i.id)
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS adl_completed_count,
  COALESCE((SELECT SUM(a.status IN (4,5))
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_adl_execution a
              ON a.tenant_id=i.tenant_id AND a.liquidation_id=-i.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',i.id)
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS adl_failed_count,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.qty ELSE 0 END)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_adl_execution a
              ON a.tenant_id=i.tenant_id AND a.liquidation_id=-i.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',i.id)
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS adl_execution_qty,
  COALESCE((SELECT SUM(CASE WHEN a.status=3 THEN a.relief_amount ELSE 0 END)
            FROM t_contract_account_liquidation_item i
            JOIN t_contract_adl_execution a
              ON a.tenant_id=i.tenant_id AND a.liquidation_id=-i.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',i.id)
            WHERE i.tenant_id=q.tenant_id AND i.account_liquidation_id=q.id),0) AS adl_execution_relief,
  COALESCE((SELECT COUNT(1)
            FROM t_contract_account_liquidation_item x
            JOIN t_contract_adl_execution a
              ON a.tenant_id=x.tenant_id AND a.liquidation_id=-x.id
             AND a.liquidation_no=CONCAT(q.liquidation_no,'-ITEM-',x.id)
            LEFT JOIN t_trade_settlement_instruction i
              ON i.tenant_id=a.tenant_id AND i.instruction_no=a.execution_no
            WHERE x.tenant_id=q.tenant_id AND x.account_liquidation_id=q.id
              AND a.asset_credit>0
              AND (i.id IS NULL OR i.status<>? OR i.reconciled_at=0)),0) AS adl_unreconciled,
  COALESCE((SELECT COUNT(1) FROM t_trade_settlement_instruction s
            WHERE s.tenant_id=q.tenant_id AND s.instruction_no=CONCAT(q.liquidation_no,'-NET')),0) AS net_instruction_count,
  COALESCE((SELECT SUM(s.status=? AND s.reconciled_at>0) FROM t_trade_settlement_instruction s
            WHERE s.tenant_id=q.tenant_id AND s.instruction_no=CONCAT(q.liquidation_no,'-NET')),0) AS net_instruction_done,
  COALESCE((SELECT COUNT(1) FROM t_trade_settlement_instruction s
            WHERE s.tenant_id=q.tenant_id AND s.instruction_no=CONCAT(q.liquidation_no,'-FEE')),0) AS fee_instruction_count,
  COALESCE((SELECT SUM(s.status=? AND s.reconciled_at>0 AND
                             s.asset_flow_no=CONCAT('PLATFORM:',s.instruction_no))
            FROM t_trade_settlement_instruction s
            WHERE s.tenant_id=q.tenant_id AND s.instruction_no=CONCAT(q.liquidation_no,'-FEE')),0) AS fee_instruction_done
FROM t_contract_account_liquidation q
WHERE q.id>? AND q.update_times<=?` + tenantClause + `
ORDER BY q.id
LIMIT ?`
	return m.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}
