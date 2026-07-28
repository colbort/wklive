package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractReconciliationCursorModel = (*customTContractReconciliationCursorModel)(nil)

type (
	// TContractReconciliationCursorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractReconciliationCursorModel.
	TContractReconciliationCursorModel interface {
		tContractReconciliationCursorModel
		FindContractOrderFillAudits(tenantID, cursor, cutoff int64, limit int) ([]*TContractReconciliationCursor, error)
		LoadReconciliationCursor(tenantID int64, checkType string, now int64) (int64, error)
		AdvanceReconciliationCursor(tenantID int64, checkType string, cursor, now int64) error
		CompleteReconciliationCycle(tenantID int64, checkType string, now int64) error
	}

	customTContractReconciliationCursorModel struct {
		*defaultTContractReconciliationCursorModel
	}
)

// NewTContractReconciliationCursorModel returns a model for the database table.
func NewTContractReconciliationCursorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractReconciliationCursorModel {
	return &customTContractReconciliationCursorModel{
		defaultTContractReconciliationCursorModel: newTContractReconciliationCursorModel(conn, c, opts...),
	}
}

// AdvanceReconciliationCursor implements [TContractReconciliationCursorModel].
func (c *customTContractReconciliationCursorModel) AdvanceReconciliationCursor(tenantID int64, checkType string, cursor int64, now int64) error {
	//	tenantClause := ""
	//	args := []any{int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), cursor, cutoff}
	//	if tenantID > 0 {
	//		tenantClause = " AND o.tenant_id=?"
	//		args = append(args, tenantID)
	//	}
	//	args = append(args, limit)
	//	query := `
	//
	// SELECT
	//
	//	o.tenant_id,
	//	o.id AS order_id,
	//	o.order_no,
	//	o.status,
	//	o.contract_value_type,
	//	s.price_scale,
	//	o.qty AS order_qty,
	//	o.filled_qty AS order_filled_qty,
	//	o.filled_amount AS order_filled_amount,
	//	o.canceled_qty AS order_canceled_qty,
	//	o.avg_price AS order_avg_price,
	//	o.fee AS order_fee,
	//	COUNT(f.id) AS fill_count,
	//	COALESCE(SUM(f.qty),0) AS fill_qty,
	//	COALESCE(SUM(f.amount),0) AS fill_amount,
	//	COALESCE(SUM(f.fee),0) AS fill_fee,
	//	COALESCE(CASE
	//	  WHEN SUM(f.qty) IS NULL OR SUM(f.qty)=0 THEN 0
	//	  WHEN o.contract_value_type=2 THEN SUM(f.qty)/SUM(f.qty/f.price)
	//	  ELSE SUM(f.price*f.qty)/SUM(f.qty)
	//	END,0) AS fill_avg_price
	//
	// FROM t_trade_order o
	// JOIN t_trade_symbol s ON s.tenant_id=o.tenant_id AND s.id=o.symbol_id
	// LEFT JOIN t_trade_fill f ON f.tenant_id=o.tenant_id AND f.order_id=o.id
	// WHERE o.product_type=? AND o.id>? AND o.update_times<=?` + tenantClause + `
	// GROUP BY o.tenant_id,o.id,o.order_no,o.status,o.contract_value_type,s.price_scale,o.qty,o.filled_qty,
	//
	//	o.filled_amount,o.canceled_qty,o.avg_price,o.fee
	//
	// ORDER BY o.id
	// LIMIT ?`
	//
	//	var rows []*contractOrderFillAudit
	//	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, args...); err != nil {
	//		return nil, err
	//	}
	//	return rows, nil
	panic("no impl")
}

// CompleteReconciliationCycle implements [TContractReconciliationCursorModel].
func (c *customTContractReconciliationCursorModel) CompleteReconciliationCycle(tenantID int64, checkType string, now int64) error {
	//	_, err := l.svcCtx.DB.ExecCtx(l.ctx, `
	//
	// INSERT INTO t_contract_reconciliation_cursor
	// (tenant_id,check_type,last_scanned_id,last_cycle_completed_at,create_times,update_times)
	// VALUES(?,?,0,0,?,?)
	// ON DUPLICATE KEY UPDATE update_times=update_times`, tenantID, checkType, now, now)
	//
	//	if err != nil {
	//		return 0, err
	//	}
	//	var cursor int64
	//	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &cursor,
	//		"SELECT last_scanned_id FROM t_contract_reconciliation_cursor WHERE tenant_id=? AND check_type=? LIMIT 1",
	//		tenantID, checkType)
	//	return cursor, err
	panic("no impl")
}

// FindContractOrderFillAudits implements [TContractReconciliationCursorModel].
func (c *customTContractReconciliationCursorModel) FindContractOrderFillAudits(tenantID int64, cursor int64, cutoff int64, limit int) ([]*TContractReconciliationCursor, error) {
	// _, err := l.svcCtx.DB.ExecCtx(l.ctx,
	// 	"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=?,update_times=? WHERE tenant_id=? AND check_type=?",
	// 	cursor, now, tenantID, checkType)
	// return err
	panic("no impl")
}

// LoadReconciliationCursor implements [TContractReconciliationCursorModel].
func (c *customTContractReconciliationCursorModel) LoadReconciliationCursor(tenantID int64, checkType string, now int64) (int64, error) {
	// _, err := l.svcCtx.DB.ExecCtx(l.ctx,
	// 	"UPDATE t_contract_reconciliation_cursor SET last_scanned_id=0,last_cycle_completed_at=?,update_times=? WHERE tenant_id=? AND check_type=?",
	// 	now, now, tenantID, checkType)
	// return err
	panic("no impl")
}
