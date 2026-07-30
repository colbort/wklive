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
	OptionOrderPageFilter struct {
		TenantId         int64
		UserId           int64
		AccountId        int64
		ContractId       int64
		UnderlyingSymbol string
		OrderNo          string
		Side             int64
		PositionEffect   int64
		OrderType        int64
		Status           int64
		Statuses         []int64
		CreateTimeStart  int64
		CreateTimeEnd    int64
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
	}

	customTOptionOrderModel struct {
		*defaultTOptionOrderModel
	}
)

func (m *defaultTOptionOrderModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionOrder, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionOrderRows, m.table)
	var item TOptionOrder
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
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
  AND NOT (user_id = ? AND account_id = ?)
  AND status IN (?, ?) AND unfilled_qty > 0 AND %s
ORDER BY %s LIMIT ? FOR UPDATE`, tOptionOrderRows, m.table, priceClause, orderBy)

	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId,
		contractId,
		side,
		excludeUserId,
		excludeAccountId,
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
  AND NOT (user_id = ? AND account_id = ?)
  AND status IN (?, ?) AND unfilled_qty > 0 AND %s
ORDER BY %s FOR UPDATE`, tOptionOrderRows, m.table, priceClause, orderBy)
	var list []*TOptionOrder
	err := m.QueryRowsNoCacheCtx(ctx, &list, query,
		tenantId, contractId, side, excludeUserId, excludeAccountId,
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		price,
	)
	return list, err
}
