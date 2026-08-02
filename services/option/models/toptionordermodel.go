package models

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
	"wklive/proto/common"
	"wklive/proto/option"
)

var _ TOptionOrderModel = (*customTOptionOrderModel)(nil)

type (
	OptionOrderBookLevel struct {
		Price           decimal.Decimal `db:"price"`
		Qty             decimal.Decimal `db:"qty"`
		OrderCount      int64           `db:"order_count"`
		ComboOrderCount int64           `db:"combo_order_count"`
	}

	OptionOrderPageFilter struct {
		TenantId             int64
		UserId               int64
		AccountId            int64
		ContractId           int64
		UnderlyingSymbol     string
		OrderNo              string
		Side                 int64
		PositionEffect       int64
		OrderType            int64
		Status               int64
		Statuses             []int64
		CreateTimeStart      int64
		CreateTimeEnd        int64
		ExcludeComboChildren bool
	}

	// TOptionOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionOrderModel.
	TOptionOrderModel interface {
		tOptionOrderModel
		FindPage(ctx context.Context, filter OptionOrderPageFilter, cursor int64, limit int64) ([]*TOptionOrder, int64, error)
		FindOneByTenantIdUserIdClientOrderId(ctx context.Context, tenantId, userId int64, clientOrderId string) (*TOptionOrder, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionOrder, error)
		FindMatchableOrders(ctx context.Context, tenantId, contractId, side, excludeUserId, excludeAccountId int64, price decimal.Decimal, limit int64) ([]*TOptionOrder, error)
		FindAllMatchableOrders(ctx context.Context, tenantId, contractId, side, excludeUserId, excludeAccountId int64, price decimal.Decimal) ([]*TOptionOrder, error)
		FindPortfolioRiskOrders(ctx context.Context, tenantId, userId, accountId int64) ([]*TOptionOrder, error)
		FindActiveCloseOrdersForUpdate(ctx context.Context, tenantId, userId, accountId, contractId int64) ([]*TOptionOrder, error)
		SumActiveOpenQty(ctx context.Context, tenantId, userId, contractId, side int64) (decimal.Decimal, error)
		FindCrossingSelfOrders(ctx context.Context, tenantId, userId, contractId, side int64, price decimal.Decimal) ([]*TOptionOrder, error)
		FindActiveMMPOrders(ctx context.Context, tenantId, userId, contractId int64, groupCode string, cursor, limit int64) ([]*TOptionOrder, error)
		FindFirstUnsafeMMPOrderForUpdate(ctx context.Context, tenantId, userId, contractId int64, groupCode string) (*TOptionOrder, error)
		HasActiveByContract(ctx context.Context, tenantId, contractId int64) (bool, error)
		HasUnsafeContractResumeOrders(ctx context.Context, tenantId, contractId int64) (bool, error)
		HasUnsafeKillSwitchReleaseOrders(ctx context.Context, tenantId, userId int64) (bool, error)
		FindOrderBookLevels(ctx context.Context, tenantId, contractId, side, limit int64) ([]*OptionOrderBookLevel, error)
		FindComboChildren(ctx context.Context, tenantId, comboOrderId int64) ([]*TOptionOrder, error)
		FindComboChildrenForUpdate(ctx context.Context, tenantId, comboOrderId int64) ([]*TOptionOrder, error)
	}

	customTOptionOrderModel struct {
		*defaultTOptionOrderModel
	}
)

// HasUnsafeContractResumeOrders includes cancellation and expiry transitions:
// the contract must remain paused until the order reaches a terminal state,
// because CANCELING/EXPIRING can still have an Asset release in flight.
func (m *defaultTOptionOrderModel) HasUnsafeContractResumeOrders(
	ctx context.Context, tenantId, contractId int64,
) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s
WHERE tenant_id=? AND contract_id=? AND status IN (?,?,?,?,?)`, m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &count, query, tenantId, contractId,
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(option.OrderStatus_ORDER_STATUS_CANCELING),
		int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
	); err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasUnsafeKillSwitchReleaseOrders reports whether a user still has an order
// that can trade or whose cancellation/expiry funds have not reached a
// terminal state. Callers serialize this check with the user's trading-control
// row so no new order can pass admission while the kill switch is active.
func (m *defaultTOptionOrderModel) HasUnsafeKillSwitchReleaseOrders(
	ctx context.Context, tenantId, userId int64,
) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s
WHERE tenant_id=? AND user_id=? AND status IN (?,?,?,?,?)`, m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &count, query, tenantId, userId,
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(option.OrderStatus_ORDER_STATUS_CANCELING),
		int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
	); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *defaultTOptionOrderModel) FindOrderBookLevels(
	ctx context.Context, tenantId, contractId, side, limit int64,
) ([]*OptionOrderBookLevel, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	orderBy := "price ASC"
	if side == int64(common.Side_SIDE_BUY) {
		orderBy = "price DESC"
	}
	query := fmt.Sprintf(`SELECT price, SUM(unfilled_qty) AS qty, COUNT(1) AS order_count,
  SUM(CASE WHEN combo_order_id>0 THEN 1 ELSE 0 END) AS combo_order_count
FROM %s
WHERE tenant_id=? AND contract_id=? AND side=? AND status IN (?,?)
  AND order_type IN (?,?) AND combo_order_id=0 AND price>0 AND unfilled_qty>0
GROUP BY price
ORDER BY %s
LIMIT ?`, m.table, orderBy)
	var items []*OptionOrderBookLevel
	err := m.QueryRowsNoCacheCtx(
		ctx, &items, query,
		tenantId, contractId, side,
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(option.OrderType_ORDER_TYPE_LIMIT),
		int64(option.OrderType_ORDER_TYPE_POST_ONLY),
		limit,
	)
	return items, err
}

func (m *defaultTOptionOrderModel) HasActiveByContract(
	ctx context.Context, tenantId, contractId int64,
) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s
WHERE tenant_id=? AND contract_id=? AND status IN (?,?,?)`, m.table)
	var count int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &count, query, tenantId, contractId,
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
	); err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindFirstUnsafeMMPOrderForUpdate locks the first MMP order that can still
// trade or whose cancellation/expiry funds have not reached a terminal state.
// MMP recovery must serialize this check with the config row so the group
// cannot be reactivated while Asset release is still pending.
func (m *defaultTOptionOrderModel) FindFirstUnsafeMMPOrderForUpdate(
	ctx context.Context, tenantId, userId, contractId int64, groupCode string,
) (*TOptionOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND contract_id = ? AND mmp = ? AND mmp_group = ?
  AND status IN (?, ?, ?, ?, ?)
ORDER BY id LIMIT 1 FOR UPDATE`, tOptionOrderRows, m.table)
	var item TOptionOrder
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query,
		tenantId, userId, contractId, int64(common.YesNo_YES_NO_YES), groupCode,
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(option.OrderStatus_ORDER_STATUS_CANCELING),
		int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionOrderModel) FindActiveMMPOrders(
	ctx context.Context, tenantId, userId, contractId int64, groupCode string, cursor, limit int64,
) ([]*TOptionOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)
	cursorClause := ""
	args := []any{
		tenantId, userId, contractId, int64(common.YesNo_YES_NO_YES), groupCode,
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
	}
	if cursor > 0 {
		cursorClause = " AND id < ?"
		args = append(args, cursor)
	}
	args = append(args, limit)
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND contract_id = ? AND mmp = ? AND mmp_group = ?
  AND status IN (?, ?, ?)%s
ORDER BY id DESC LIMIT ?`, tOptionOrderRows, m.table, cursorClause)
	var items []*TOptionOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *defaultTOptionOrderModel) SumActiveOpenQty(
	ctx context.Context, tenantId, userId, contractId, side int64,
) (decimal.Decimal, error) {
	userClause := ""
	args := []any{
		tenantId, contractId, side,
		int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
	}
	if userId > 0 {
		userClause = " AND user_id = ?"
		args = append(args, userId)
	}
	query := fmt.Sprintf(`SELECT COALESCE(SUM(unfilled_qty), 0) AS total FROM %s
WHERE tenant_id = ? AND contract_id = ? AND side = ? AND position_effect = ?
  AND status IN (?, ?, ?)%s`, m.table, userClause)
	var aggregate decimalAggregate
	if err := m.QueryRowNoCacheCtx(ctx, &aggregate, query, args...); err != nil {
		return decimal.Zero, err
	}
	return aggregate.Decimal()
}

func (m *defaultTOptionOrderModel) FindCrossingSelfOrders(
	ctx context.Context, tenantId, userId, contractId, side int64, price decimal.Decimal,
) ([]*TOptionOrder, error) {
	priceClause := "price <= ?"
	orderBy := "price ASC, id ASC"
	if side == int64(common.Side_SIDE_BUY) {
		priceClause = "price >= ?"
		orderBy = "price DESC, id ASC"
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND contract_id = ? AND side = ?
  AND combo_order_id = 0 AND status IN (?, ?) AND unfilled_qty > 0 AND %s
ORDER BY %s FOR UPDATE`, tOptionOrderRows, m.table, priceClause, orderBy)
	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(
		ctx, &list, query,
		tenantId, userId, contractId, side,
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		price,
	)
	return list, err
}

func (m *defaultTOptionOrderModel) FindActiveCloseOrdersForUpdate(
	ctx context.Context,
	tenantId, userId, accountId, contractId int64,
) ([]*TOptionOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND account_id = ? AND contract_id = ?
  AND side = ? AND position_effect = ? AND status IN (?, ?, ?)
ORDER BY id FOR UPDATE`, tOptionOrderRows, m.table)
	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(
		ctx, &list, query,
		tenantId, userId, accountId, contractId,
		int64(common.Side_SIDE_BUY),
		int64(option.PositionEffect_POSITION_EFFECT_CLOSE),
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
	)
	return list, err
}

func (m *defaultTOptionOrderModel) FindPortfolioRiskOrders(
	ctx context.Context,
	tenantId, userId, accountId int64,
) ([]*TOptionOrder, error) {
	accountClause := ""
	args := []any{tenantId, userId}
	if accountId > 0 {
		accountClause = " AND account_id = ?"
		args = append(args, accountId)
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ?%s AND side = ?
  AND status IN (?,?,?,?,?) AND unfilled_qty > 0
ORDER BY id FOR UPDATE`, tOptionOrderRows, m.table, accountClause)
	var list []*TOptionOrder
	args = append(args,
		int64(common.Side_SIDE_SELL),
		int64(option.OrderStatus_ORDER_STATUS_FUNDING),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		int64(option.OrderStatus_ORDER_STATUS_CANCELING),
		int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
	)
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	return list, err
}

func (m *defaultTOptionOrderModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionOrder, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionOrderRows, m.table)
	var item TOptionOrder
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionOrderModel) FindComboChildren(
	ctx context.Context, tenantId, comboOrderId int64,
) ([]*TOptionOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND combo_order_id = ?
ORDER BY combo_leg_no, id`, tOptionOrderRows, m.table)
	var list []*TOptionOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTOptionOrderModel) FindComboChildrenForUpdate(
	ctx context.Context, tenantId, comboOrderId int64,
) ([]*TOptionOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND combo_order_id = ?
ORDER BY combo_leg_no, id FOR UPDATE`, tOptionOrderRows, m.table)
	var list []*TOptionOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, comboOrderId); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTOptionOrderModel) FindOneByTenantIdUserIdClientOrderId(ctx context.Context, tenantId, userId int64, clientOrderId string) (*TOptionOrder, error) {
	if clientOrderId == "" {
		return nil, ErrNotFound
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND user_id = ? AND client_order_id = ? LIMIT 1", tOptionOrderRows, m.table)
	var item TOptionOrder
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, userId, clientOrderId); err != nil {
		return nil, err
	}
	return &item, nil
}

// NewTOptionOrderModel returns a model for the database table.
func NewTOptionOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionOrderModel {
	return &customTOptionOrderModel{
		defaultTOptionOrderModel: newTOptionOrderModel(conn, c, opts...),
	}
}

func (m *defaultTOptionOrderModel) FindPage(ctx context.Context, filter OptionOrderPageFilter, cursor int64, limit int64) ([]*TOptionOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("position_effect", filter.PositionEffect)
	builder.EqInt64("order_type", filter.OrderType)
	builder.EqInt64("status", filter.Status)
	builder.InInt64("status", filter.Statuses)
	builder.GteInt64("create_times", filter.CreateTimeStart)
	builder.LteInt64("create_times", filter.CreateTimeEnd)
	if filter.ExcludeComboChildren {
		builder.And("combo_order_id = 0")
	}

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionOrderRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTOptionOrderModel) FindMatchableOrders(ctx context.Context, tenantId, contractId, side, excludeUserId, excludeAccountId int64, price decimal.Decimal, limit int64) ([]*TOptionOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)

	priceClause := "price <= ?"
	orderBy := "price ASC, id ASC"
	if side == 1 {
		priceClause = "price >= ?"
		orderBy = "price DESC, id ASC"
	}

	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND contract_id = ? AND side = ?
  AND combo_order_id = 0 AND user_id <> ?
  AND status IN (?, ?) AND unfilled_qty > 0 AND %s
ORDER BY %s LIMIT ? FOR UPDATE`, tOptionOrderRows, m.table, priceClause, orderBy)

	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId,
		contractId,
		side,
		excludeUserId,
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		price,
		limit,
	)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultTOptionOrderModel) FindAllMatchableOrders(ctx context.Context, tenantId, contractId, side, excludeUserId, excludeAccountId int64, price decimal.Decimal) ([]*TOptionOrder, error) {
	priceClause := "price <= ?"
	orderBy := "price ASC, id ASC"
	if side == int64(common.Side_SIDE_BUY) {
		priceClause = "price >= ?"
		orderBy = "price DESC, id ASC"
	}
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND contract_id = ? AND side = ?
  AND combo_order_id = 0 AND user_id <> ?
  AND status IN (?, ?) AND unfilled_qty > 0 AND %s
ORDER BY %s FOR UPDATE`, tOptionOrderRows, m.table, priceClause, orderBy)
	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId, contractId, side, excludeUserId,
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		price,
	)
	return list, err
}
