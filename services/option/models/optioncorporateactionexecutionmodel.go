package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CountCorporateActionExecutionBlockers returns unfinished records that make a
// source contract unsafe to migrate during a corporate action.
func CountCorporateActionExecutionBlockers(
	ctx context.Context, conn sqlx.SqlConn, tenantID, contractID int64,
) (int64, error) {
	const query = `SELECT
  (SELECT COUNT(1) FROM t_option_order
    WHERE tenant_id=? AND contract_id=? AND status IN (1,2,7,8,9))
 + (SELECT COUNT(1) FROM t_option_outbox
    WHERE tenant_id=? AND contract_id=? AND status<>3)
 + (SELECT COUNT(1) FROM t_option_inbox
    WHERE tenant_id=? AND contract_id=? AND status<>2)
 + (SELECT COUNT(1) FROM t_option_asset_instruction ai
    LEFT JOIN t_option_order ai_order
      ON ai_order.tenant_id=ai.tenant_id AND ai_order.id=ai.order_id
    LEFT JOIN t_option_trade ai_trade
      ON ai_trade.tenant_id=ai.tenant_id AND ai_trade.id=ai.trade_id
    LEFT JOIN t_option_position ai_position
      ON ai_position.tenant_id=ai.tenant_id AND ai_position.id=ai.position_id
    LEFT JOIN t_option_margin_lot ai_margin_lot
      ON ai_margin_lot.tenant_id=ai.tenant_id AND ai_margin_lot.id=ai.margin_lot_id
    LEFT JOIN t_option_liquidation ai_liquidation
      ON ai_liquidation.tenant_id=ai.tenant_id AND ai_liquidation.id=ai.liquidation_id
    LEFT JOIN t_option_physical_delivery_unit ai_delivery_unit
      ON ai_delivery_unit.tenant_id=ai.tenant_id AND ai_delivery_unit.id=ai.delivery_unit_id
    WHERE ai.tenant_id=? AND ai.status NOT IN (3,6)
      AND ? IN (
        ai_order.contract_id, ai_trade.contract_id, ai_position.contract_id,
        ai_margin_lot.contract_id, ai_liquidation.contract_id, ai_delivery_unit.contract_id
      ))
 + (SELECT COUNT(1) FROM t_option_exercise
    WHERE tenant_id=? AND contract_id=? AND status=1)
 + (SELECT COUNT(1) FROM t_option_liquidation
    WHERE tenant_id=? AND contract_id=? AND status IN (1,2,4,6))
 + (SELECT COUNT(1) FROM t_option_settlement_batch
    WHERE tenant_id=? AND contract_id=? AND status<>7)
 + (SELECT COUNT(1) FROM t_option_physical_delivery_unit
    WHERE tenant_id=? AND contract_id=? AND status NOT IN (5,6)) AS blocked`
	args := make([]any, 0, 16)
	for index := 0; index < 8; index++ {
		args = append(args, tenantID, contractID)
	}
	var blocked int64
	if err := conn.QueryRowCtx(ctx, &blocked, query, args...); err != nil {
		return 0, err
	}
	return blocked, nil
}
