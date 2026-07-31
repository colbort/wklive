package models

import (
	"context"
	"fmt"

	"wklive/proto/common"
	"wklive/proto/option"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionOutboxModel = (*customTOptionOutboxModel)(nil)

type (
	// TOptionOutboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionOutboxModel.
	TOptionOutboxModel interface {
		tOptionOutboxModel
		FindRunnable(ctx context.Context, tenantId, now, limit int64) ([]*TOptionOutbox, error)
		ComboDebitBarrierReady(ctx context.Context, tenantId int64, comboMatchNo string) (bool, error)
		CountStaleComboDebitBarrierBlocked(ctx context.Context, tenantId, staleBefore int64) (int64, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionOutbox, error)
		Claim(ctx context.Context, id, now int64) (bool, error)
		RecoverStale(ctx context.Context, staleBefore, now int64) error
		HasIncomplete(ctx context.Context, tenantId, contractId int64) (bool, error)
		ResetForManualRetry(ctx context.Context, id, now int64) (bool, error)
		SumPendingPositionDelta(ctx context.Context, tenantId, userId, contractId, positionSide int64) (decimal.Decimal, error)
	}

	customTOptionOutboxModel struct {
		*defaultTOptionOutboxModel
	}
)

type comboDebitBarrierState struct {
	TradeLegs     int64 `db:"trade_legs"`
	DistinctLegs  int64 `db:"distinct_legs"`
	MinLeg        int64 `db:"min_leg"`
	MaxLeg        int64 `db:"max_leg"`
	MissingDebits int64 `db:"missing_debits"`
}

// SumPendingPositionDelta bridges the short interval between a match updating
// order.unfilled_qty and its outbox event updating the position. The position
// update and outbox SUCCESS transition commit in one transaction, so exactly
// one of the persisted position or this delta represents each fill.
func (m *defaultTOptionOutboxModel) SumPendingPositionDelta(
	ctx context.Context, tenantId, userId, contractId, positionSide int64,
) (decimal.Decimal, error) {
	type participant struct {
		orderAlias string
		orderID    string
		effect     int64
		sign       int64
	}
	var participants []participant
	switch positionSide {
	case int64(common.PositionSide_POSITION_SIDE_LONG):
		participants = []participant{
			{orderAlias: "o", orderID: "t.buy_order_id", effect: int64(option.PositionEffect_POSITION_EFFECT_OPEN), sign: 1},
			{orderAlias: "o", orderID: "t.sell_order_id", effect: int64(option.PositionEffect_POSITION_EFFECT_CLOSE), sign: -1},
		}
	case int64(common.PositionSide_POSITION_SIDE_SHORT):
		participants = []participant{
			{orderAlias: "o", orderID: "t.sell_order_id", effect: int64(option.PositionEffect_POSITION_EFFECT_OPEN), sign: 1},
			{orderAlias: "o", orderID: "t.buy_order_id", effect: int64(option.PositionEffect_POSITION_EFFECT_CLOSE), sign: -1},
		}
	default:
		return decimal.Zero, nil
	}
	total := decimal.Zero
	for _, p := range participants {
		query := fmt.Sprintf(`SELECT COALESCE(SUM(t.qty), 0) AS total
FROM %s e
JOIN t_option_trade t ON t.id = e.trade_id AND t.tenant_id = e.tenant_id
JOIN t_option_order %s ON %s.id = %s AND %s.tenant_id = e.tenant_id
WHERE e.tenant_id = ? AND e.contract_id = ? AND e.event_type = ?
  AND e.status <> ? AND %s.position_effect = ?
  AND (? = 0 OR %s.user_id = ?)`,
			m.table, p.orderAlias, p.orderAlias, p.orderID, p.orderAlias, p.orderAlias, p.orderAlias)
		var aggregate decimalAggregate
		if err := m.QueryRowNoCacheCtx(
			ctx, &aggregate, query,
			tenantId, contractId,
			int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
			p.effect, userId, userId,
		); err != nil {
			return decimal.Zero, err
		}
		amount, err := aggregate.Decimal()
		if err != nil {
			return decimal.Zero, err
		}
		if p.sign > 0 {
			total = total.Add(amount)
		} else {
			total = total.Sub(amount)
		}
	}
	return total, nil
}

func (m *defaultTOptionOutboxModel) FindRunnable(ctx context.Context, tenantId, now, limit int64) ([]*TOptionOutbox, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := fmt.Sprintf(`SELECT %s FROM %s AS current
WHERE (? = 0 OR current.tenant_id = ?)
AND current.status IN (?, ?) AND current.next_retry_at <= ?
AND EXISTS (
  SELECT 1 FROM t_option_asset_instruction premium_debit
  WHERE premium_debit.tenant_id = current.tenant_id
    AND premium_debit.trade_id = current.trade_id
    AND premium_debit.action = ?
    AND premium_debit.step_no = 1
    AND premium_debit.status = ?
)
AND NOT EXISTS (
  SELECT 1
  FROM t_option_trade combo_current
  JOIN t_option_trade combo_sibling
    ON combo_sibling.tenant_id = combo_current.tenant_id
   AND combo_sibling.combo_match_no = combo_current.combo_match_no
  WHERE combo_current.id = current.trade_id
    AND combo_current.tenant_id = current.tenant_id
    AND combo_current.combo_match_no <> ''
    AND NOT EXISTS (
      SELECT 1 FROM t_option_asset_instruction sibling_debit
      WHERE sibling_debit.tenant_id = combo_sibling.tenant_id
        AND sibling_debit.trade_id = combo_sibling.id
        AND sibling_debit.action = ?
        AND sibling_debit.step_no = 1
        AND sibling_debit.status = ?
    )
)
AND NOT EXISTS (
  SELECT 1 FROM %s AS previous
  WHERE previous.tenant_id = current.tenant_id
    AND previous.contract_id = current.contract_id
    AND previous.event_type = current.event_type
    AND previous.match_sequence < current.match_sequence
    AND previous.status <> ?
)
ORDER BY current.id LIMIT ?`, tOptionOutboxRows, m.table, m.table)
	var list []*TOptionOutbox
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId, tenantId,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
		now,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
		limit,
	)
	return list, err
}

func (m *defaultTOptionOutboxModel) ComboDebitBarrierReady(
	ctx context.Context, tenantId int64, comboMatchNo string,
) (bool, error) {
	if comboMatchNo == "" {
		return true, nil
	}
	query := `SELECT
  COUNT(1) trade_legs,
  COUNT(DISTINCT sibling.combo_leg_no) distinct_legs,
  COALESCE(MIN(sibling.combo_leg_no),0) min_leg,
  COALESCE(MAX(sibling.combo_leg_no),0) max_leg,
  COALESCE(SUM(CASE WHEN NOT EXISTS (
    SELECT 1 FROM t_option_asset_instruction debit
    WHERE debit.tenant_id=sibling.tenant_id
      AND debit.trade_id=sibling.id
      AND debit.action=?
      AND debit.step_no=1
      AND debit.status=?
  ) THEN 1 ELSE 0 END),0) missing_debits
FROM t_option_trade sibling
WHERE sibling.tenant_id=? AND sibling.combo_match_no=?`
	var state comboDebitBarrierState
	if err := m.QueryRowNoCacheCtx(
		ctx, &state, query,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		tenantId, comboMatchNo,
	); err != nil {
		return false, err
	}
	return state.TradeLegs >= 2 &&
		state.TradeLegs <= 4 &&
		state.TradeLegs == state.DistinctLegs &&
		state.MinLeg == 1 &&
		state.MaxLeg == state.TradeLegs &&
		state.MissingDebits == 0, nil
}

func (m *defaultTOptionOutboxModel) CountStaleComboDebitBarrierBlocked(
	ctx context.Context, tenantId, staleBefore int64,
) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(1)
FROM %s current
JOIN t_option_trade combo_current
  ON combo_current.tenant_id=current.tenant_id
 AND combo_current.id=current.trade_id
WHERE (?=0 OR current.tenant_id=?)
  AND current.event_type=?
  AND current.status IN (?,?,?)
  AND current.create_times<?
  AND combo_current.combo_match_no<>''
  AND EXISTS (
    SELECT 1
    FROM t_option_trade sibling
    WHERE sibling.tenant_id=combo_current.tenant_id
      AND sibling.combo_match_no=combo_current.combo_match_no
      AND NOT EXISTS (
        SELECT 1 FROM t_option_asset_instruction debit
        WHERE debit.tenant_id=sibling.tenant_id
          AND debit.trade_id=sibling.id
          AND debit.action=?
          AND debit.step_no=1
          AND debit.status=?
      )
  )`, m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &count, query,
		tenantId, tenantId,
		int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_MANUAL_REVIEW),
		staleBefore,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
	); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultTOptionOutboxModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionOutbox, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionOutboxRows, m.table)
	var item TOptionOutbox
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionOutboxModel) Claim(ctx context.Context, id, now int64) (bool, error) {
	query := fmt.Sprintf("UPDATE %s SET status = ?, update_times = ? WHERE id = ? AND status IN (?, ?)", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), now, id,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func (m *defaultTOptionOutboxModel) RecoverStale(ctx context.Context, staleBefore, now int64) error {
	query := fmt.Sprintf(`UPDATE %s SET status = ?, next_retry_at = ?, update_times = ?,
last_error_msg = 'recovered stale processing event'
WHERE status = ? AND update_times < ?`, m.table)
	_, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED), now, now,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING), staleBefore,
	)
	return err
}

func (m *defaultTOptionOutboxModel) HasIncomplete(ctx context.Context, tenantId, contractId int64) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND contract_id = ? AND status <> ?", m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantId, contractId, int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS)); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *defaultTOptionOutboxModel) ResetForManualRetry(ctx context.Context, id, now int64) (bool, error) {
	query := fmt.Sprintf(`UPDATE %s SET status = ?, retry_count = 0, next_retry_at = ?,
last_error_msg = '', update_times = ? WHERE id = ? AND status IN (?, ?)`, m.table)
	result, err := m.ExecNoCacheCtx(ctx, query,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING), now, now, id,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_MANUAL_REVIEW),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

// NewTOptionOutboxModel returns a model for the database table.
func NewTOptionOutboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionOutboxModel {
	return &customTOptionOutboxModel{
		defaultTOptionOutboxModel: newTOptionOutboxModel(conn, c, opts...),
	}
}
