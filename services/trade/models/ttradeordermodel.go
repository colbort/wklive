package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeOrderModel = (*customTTradeOrderModel)(nil)

type (
	TradeOrderPageFilter struct {
		TenantId        int64
		UserId          int64
		SymbolId        int64
		ProductType     int64
		Status          int64
		Side            int64
		TimeStart       int64
		TimeEnd         int64
		Keyword         string
		Statuses        []int64
		ExcludeStatuses []int64
		ExcludeSources  []int64
		PositionSide    int64
	}

	TradeOrderMatchKey struct {
		TenantId    int64 `db:"tenant_id"`
		SymbolId    int64 `db:"symbol_id"`
		ProductType int64 `db:"product_type"`
	}

	// TTradeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderModel.
	TTradeOrderModel interface {
		tTradeOrderModel
		FindPage(ctx context.Context, filter TradeOrderPageFilter, cursor int64, limit int64) ([]*TTradeOrder, int64, error)
		CountByStatuses(ctx context.Context, tenantId, userId uint64, marketType int64, statuses []int64) (int64, error)
		FindMatchKeys(ctx context.Context, tenantId int64, statuses []int64, limit int64) ([]TradeOrderMatchKey, error)
		FindOpenMatchOrders(ctx context.Context, tenantId, symbolId, marketType, side int64, statuses []int64, marketOrderType int64, limit int64) ([]*TTradeOrder, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrder, error)
		FindOneByTenantIdOrderNoForUpdate(ctx context.Context, tenantId int64, orderNo string) (*TTradeOrder, error)
		FindOneByTenantIdUserIdClientOrderId(ctx context.Context, tenantId, userId int64, clientOrderId sql.NullString) (*TTradeOrder, error)
		CountBySymbolStatuses(ctx context.Context, tenantID, symbolID int64, statuses []int64) (int64, error)
		CountByUserSymbolStatuses(ctx context.Context, tenantID, userID, symbolID int64, statuses []int64) (int64, error)
		CountCreatedSince(ctx context.Context, tenantID, userID, productType, contractType, since int64) (int64, error)
		CountCancelsSince(ctx context.Context, tenantID, userID, productType, contractType, since int64) (int64, error)
		CountOpenContractRiskUnit(ctx context.Context, tenantID, userID, symbolID int64) (int64, error)
		CountActiveIncompatibleContractMode(ctx context.Context, tenantID, userID, symbolID, marginMode, positionMode int64) (int64, error)
		CountFreezingCrossMarginOpenings(ctx context.Context, tenantID, userID int64, marginAsset string) (int64, error)
		FindCrossMarginCancelableOrderIDs(ctx context.Context, tenantID, userID int64, marginAsset string) ([]int64, error)
		FindTerminalAssetRepairCandidates(ctx context.Context, tenantID, cursor, limit int64, statuses []int64) ([]*TTradeOrder, error)
		ArchiveZeroFillLiquidityOrders(ctx context.Context, source, canceledStatus, cutoff, batchSize, archivedAt int64) (int64, error)
	}

	customTTradeOrderModel struct {
		*defaultTTradeOrderModel
	}
)

func (m *customTTradeOrderModel) CountByUserSymbolStatuses(
	ctx context.Context, tenantID, userID, symbolID int64, statuses []int64,
) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	holders := make([]string, len(statuses))
	args := []any{tenantID, userID, symbolID}
	for i, status := range statuses {
		holders[i] = "?"
		args = append(args, status)
	}
	var total int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id=? AND user_id=? AND symbol_id=? AND status IN (%s)", m.table, strings.Join(holders, ","))
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customTTradeOrderModel) CountCreatedSince(
	ctx context.Context, tenantID, userID, productType, contractType, since int64,
) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id=? AND user_id=? AND product_type=? AND create_times>=?", m.table)
	args := []any{tenantID, userID, productType, since}
	if productType == 2 && contractType > 0 {
		query += " AND contract_type=?"
		args = append(args, contractType)
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customTTradeOrderModel) CountCancelsSince(
	ctx context.Context, tenantID, userID, productType, contractType, since int64,
) (int64, error) {
	query := `SELECT COUNT(1)
FROM t_trade_cancel_log c
JOIN t_trade_order o ON o.tenant_id=c.tenant_id AND o.id=c.order_id
WHERE c.tenant_id=? AND c.user_id=? AND o.product_type=? AND c.create_times>=?`
	args := []any{tenantID, userID, productType, since}
	if productType == 2 && contractType > 0 {
		query += " AND o.contract_type=?"
		args = append(args, contractType)
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customTTradeOrderModel) CountOpenContractRiskUnit(
	ctx context.Context, tenantID, userID, symbolID int64,
) (int64, error) {
	query := `SELECT COUNT(1) FROM t_trade_order
WHERE tenant_id=? AND user_id=? AND product_type=2
  AND status IN (1,2,7,8,9,10,11)`
	args := []any{tenantID, userID}
	if symbolID > 0 {
		query += " AND symbol_id=?"
		args = append(args, symbolID)
	}
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customTTradeOrderModel) CountActiveIncompatibleContractMode(
	ctx context.Context, tenantID, userID, symbolID, marginMode, positionMode int64,
) (int64, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_trade_order o
JOIN t_trade_order_contract c
  ON c.tenant_id=o.tenant_id AND c.order_id=o.id
WHERE o.tenant_id=? AND o.user_id=? AND o.symbol_id=?
  AND o.status IN (1,2,7,8,9,10,11)
  AND (
    c.margin_mode<>? OR
    (?=1 AND o.position_side<>3) OR
    (?=2 AND o.position_side=3)
  )`,
		tenantID, userID, symbolID, marginMode, positionMode, positionMode)
	return count, err
}

func (m *customTTradeOrderModel) CountFreezingCrossMarginOpenings(
	ctx context.Context, tenantID, userID int64, marginAsset string,
) (int64, error) {
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, `SELECT COUNT(1)
FROM t_trade_order o
JOIN t_trade_order_contract c
  ON c.tenant_id=o.tenant_id AND c.order_id=o.id
JOIN t_trade_asset_reservation r
  ON r.tenant_id=o.tenant_id AND r.order_id=o.id
WHERE o.tenant_id=? AND o.user_id=? AND c.margin_asset=?
  AND c.margin_mode=1 AND c.margin_amount>0
  AND o.status=7 AND r.status=1`,
		tenantID, userID, marginAsset)
	return count, err
}

func (m *customTTradeOrderModel) FindCrossMarginCancelableOrderIDs(
	ctx context.Context, tenantID, userID int64, marginAsset string,
) ([]int64, error) {
	type row struct {
		ID int64 `db:"id"`
	}
	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, `SELECT o.id
FROM t_trade_order o
JOIN t_trade_order_contract c
  ON c.tenant_id=o.tenant_id AND c.order_id=o.id
WHERE o.tenant_id=? AND o.user_id=? AND c.margin_mode=1 AND c.margin_asset=?
  AND o.status IN (1,2,8)
ORDER BY o.id`, tenantID, userID, marginAsset); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, item := range rows {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

// FindTerminalAssetRepairCandidates returns only terminal orders whose durable
// reservation still has a positive unsettled remainder. Scanning every
// historical terminal order makes the event recovery task monopolize its
// distributed lock after a large database import.
func (m *customTTradeOrderModel) FindTerminalAssetRepairCandidates(
	ctx context.Context,
	tenantID, cursor, limit int64,
	statuses []int64,
) ([]*TTradeOrder, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	limit = sqlutil.NormalizeLimit(limit)
	holders := make([]string, len(statuses))
	args := make([]any, 0, len(statuses)+4)
	for i, status := range statuses {
		holders[i] = "?"
		args = append(args, status)
	}
	query := fmt.Sprintf(
		`SELECT %s FROM %s o
WHERE o.status IN (%s)
  AND o.id>?
  AND EXISTS (
    SELECT 1 FROM t_trade_asset_reservation r
    WHERE r.tenant_id=o.tenant_id AND r.order_id=o.id
      AND r.reserved_amount>r.consumed_amount+r.released_amount
  )`,
		tTradeOrderRows,
		m.table,
		strings.Join(holders, ","),
	)
	args = append(args, cursor)
	if tenantID > 0 {
		query += " AND o.tenant_id=?"
		args = append(args, tenantID)
	}
	query += " ORDER BY o.id ASC LIMIT ?"
	args = append(args, limit)
	var rows []*TTradeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// NewTTradeOrderModel returns a model for the database table.
func NewTTradeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderModel {
	return &customTTradeOrderModel{
		defaultTTradeOrderModel: newTTradeOrderModel(conn, c, opts...),
	}
}

func (m *customTTradeOrderModel) ArchiveZeroFillLiquidityOrders(ctx context.Context, source, canceledStatus, cutoff, batchSize, archivedAt int64) (int64, error) {
	if cutoff <= 0 || batchSize <= 0 {
		return 0, nil
	}
	if batchSize > 1000 {
		batchSize = 1000
	}

	var rows []*TTradeOrder
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		query := fmt.Sprintf(
			"SELECT %s FROM %s o WHERE o.source=? AND o.status=? AND o.filled_qty=0 AND o.filled_amount=0 AND o.update_times>0 AND o.update_times<? AND NOT EXISTS (SELECT 1 FROM `t_trade_fill` f WHERE f.tenant_id=o.tenant_id AND f.order_id=o.id) ORDER BY o.id LIMIT ? FOR UPDATE",
			tTradeOrderRows,
			m.table,
		)
		if err := session.QueryRowsCtx(ctx, &rows, query, source, canceledStatus, cutoff, batchSize); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}

		parts := make([]string, 0, len(rows))
		idArgs := make([]any, 0, len(rows))
		for _, row := range rows {
			parts = append(parts, "?")
			idArgs = append(idArgs, row.Id)
		}
		inClause := strings.Join(parts, ",")
		archiveArgs := append([]any{archivedAt, "ZERO_FILL_CANCELED"}, idArgs...)

		for _, table := range []string{"t_trade_order_spot", "t_trade_order_contract", "t_trade_cancel_log"} {
			archiveTable := table + "_archive"
			stmt := fmt.Sprintf(
				"INSERT IGNORE INTO `%s` SELECT c.*, ?, ? FROM `%s` c WHERE c.order_id IN (%s)",
				archiveTable,
				table,
				inClause,
			)
			args := append([]any{archivedAt, "ZERO_FILL_CANCELED"}, idArgs...)
			if _, err := session.ExecCtx(ctx, stmt, args...); err != nil {
				return err
			}
		}
		if _, err := session.ExecCtx(ctx,
			fmt.Sprintf("INSERT IGNORE INTO `t_trade_order_archive` SELECT o.*, ?, ? FROM `t_trade_order` o WHERE o.id IN (%s)", inClause),
			archiveArgs...,
		); err != nil {
			return err
		}
		for _, table := range []string{"t_trade_order_spot", "t_trade_order_contract", "t_trade_cancel_log"} {
			if _, err := session.ExecCtx(ctx, fmt.Sprintf("DELETE FROM `%s` WHERE order_id IN (%s)", table, inClause), idArgs...); err != nil {
				return err
			}
		}
		_, err := session.ExecCtx(ctx, fmt.Sprintf("DELETE FROM `t_trade_order` WHERE id IN (%s)", inClause), idArgs...)
		return err
	})
	if err != nil {
		return 0, err
	}

	keys := make([]string, 0, len(rows)*3)
	for _, row := range rows {
		keys = append(keys,
			fmt.Sprintf("%s%v", cacheTTradeOrderIdPrefix, row.Id),
			fmt.Sprintf("%s%v:%v", cacheTTradeOrderTenantIdOrderNoPrefix, row.TenantId, row.OrderNo),
			fmt.Sprintf("%s%v:%v:%v:%v", cacheTTradeOrderTenantIdUserIdProductTypeClientOrderIdPrefix, row.TenantId, row.UserId, row.ProductType, row.ClientOrderId),
		)
	}
	if len(keys) > 0 {
		if err := m.DelCacheCtx(ctx, keys...); err != nil {
			return 0, err
		}
	}
	return int64(len(rows)), nil
}

func (m *customTTradeOrderModel) CountBySymbolStatuses(ctx context.Context, tenantID, symbolID int64, statuses []int64) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	holders := make([]string, len(statuses))
	args := []any{tenantID, symbolID}
	for i, status := range statuses {
		holders[i] = "?"
		args = append(args, status)
	}
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id=? AND symbol_id=? AND status IN (%s)", m.table, strings.Join(holders, ","))
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customTTradeOrderModel) FindPage(ctx context.Context, filter TradeOrderPageFilter, cursor int64, limit int64) ([]*TTradeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("product_type", filter.ProductType)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("position_side", filter.PositionSide)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)
	builder.InInt64("status", filter.Statuses)
	if len(filter.ExcludeStatuses) > 0 {
		holders := make([]any, 0, len(filter.ExcludeStatuses))
		parts := make([]string, 0, len(filter.ExcludeStatuses))
		for _, item := range filter.ExcludeStatuses {
			parts = append(parts, "?")
			holders = append(holders, item)
		}
		builder.And(fmt.Sprintf("status NOT IN (%s)", joinComma(parts)), holders...)
	}
	if len(filter.ExcludeSources) > 0 {
		holders := make([]any, 0, len(filter.ExcludeSources))
		parts := make([]string, 0, len(filter.ExcludeSources))
		for _, item := range filter.ExcludeSources {
			parts = append(parts, "?")
			holders = append(holders, item)
		}
		builder.And(fmt.Sprintf("source NOT IN (%s)", joinComma(parts)), holders...)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		builder.And("(order_no LIKE ? OR client_order_id LIKE ?)", kw, kw)
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tTradeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TTradeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *customTTradeOrderModel) CountByStatuses(ctx context.Context, tenantId, userId uint64, marketType int64, statuses []int64) (int64, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", int64(tenantId))
	builder.EqInt64("user_id", int64(userId))
	builder.EqInt64("product_type", marketType)
	builder.InInt64("status", statuses)

	var total int64
	sql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, builder.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &total, sql, builder.Args()...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customTTradeOrderModel) FindMatchKeys(ctx context.Context, tenantId int64, statuses []int64, limit int64) ([]TradeOrderMatchKey, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where, args := openOrderWhere(tenantId, 0, 0, 0, statuses)
	sql := fmt.Sprintf("SELECT tenant_id, symbol_id, product_type FROM %s WHERE %s AND order_type IN (?, ?) GROUP BY tenant_id, symbol_id, product_type ORDER BY tenant_id ASC, symbol_id ASC, product_type ASC LIMIT ?", m.table, where)
	args = append(args, 1, 2, limit)

	var list []TradeOrderMatchKey
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTTradeOrderModel) FindOpenMatchOrders(ctx context.Context, tenantId, symbolId, marketType, side int64, statuses []int64, marketOrderType int64, limit int64) ([]*TTradeOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)
	where, args := openOrderWhere(tenantId, symbolId, marketType, side, statuses)

	priceOrder := "price ASC"
	if side == 1 {
		priceOrder = "price DESC"
	}
	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s AND order_type IN (?, ?) ORDER BY CASE WHEN order_type = ? THEN 0 ELSE 1 END ASC, %s, id ASC LIMIT ?",
		tTradeOrderRows,
		m.table,
		where,
		priceOrder,
	)
	args = append(args, 1, 2, marketOrderType, limit)

	var list []*TTradeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTTradeOrderModel) FindOneForUpdate(ctx context.Context, id int64) (*TTradeOrder, error) {
	var resp TTradeOrder
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1 FOR UPDATE", tTradeOrderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, sql, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customTTradeOrderModel) FindOneByTenantIdOrderNoForUpdate(ctx context.Context, tenantId int64, orderNo string) (*TTradeOrder, error) {
	var resp TTradeOrder
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE `tenant_id` = ? AND `order_no` = ? LIMIT 1 FOR UPDATE", tTradeOrderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, sql, tenantId, orderNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOneByTenantIdUserIdClientOrderId supports endpoints that do not carry product_type,
// such as cancel-by-client-order-id. New order idempotency checks should use the generated
// product-scoped unique-key lookup.
func (m *customTTradeOrderModel) FindOneByTenantIdUserIdClientOrderId(ctx context.Context, tenantId, userId int64, clientOrderId sql.NullString) (*TTradeOrder, error) {
	var resp TTradeOrder
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `tenant_id` = ? AND `user_id` = ? AND `client_order_id` = ? ORDER BY `id` DESC LIMIT 1", tTradeOrderRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantId, userId, clientOrderId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func openOrderWhere(tenantId, symbolId, marketType, side int64, statuses []int64) (string, []any) {
	parts := make([]string, 0, 5)
	args := make([]any, 0, 8)
	if tenantId > 0 {
		parts = append(parts, "tenant_id = ?")
		args = append(args, tenantId)
	}
	if symbolId > 0 {
		parts = append(parts, "symbol_id = ?")
		args = append(args, symbolId)
	}
	if marketType > 0 {
		parts = append(parts, "product_type = ?")
		args = append(args, marketType)
	}
	if side > 0 {
		parts = append(parts, "side = ?")
		args = append(args, side)
	}
	if len(statuses) > 0 {
		holders := make([]string, 0, len(statuses))
		for _, status := range statuses {
			holders = append(holders, "?")
			args = append(args, status)
		}
		parts = append(parts, fmt.Sprintf("status IN (%s)", joinComma(holders)))
	}
	if len(parts) == 0 {
		return "1=1", args
	}
	return strings.Join(parts, " AND "), args
}

func joinComma(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += "," + items[i]
	}
	return out
}
