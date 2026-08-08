package models

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeOrderModel = (*customTStakeOrderModel)(nil)

type (
	StakeOrderPageFilter struct {
		TenantId    int64
		UserId      int64
		ProductId   int64
		OrderNo     string
		ProductName string
		CoinSymbol  string
		Status      int64
		RedeemType  int64
		Source      int64
		StartBegin  int64
		StartEnd    int64
		EndBegin    int64
		EndEnd      int64
	}

	// TStakeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeOrderModel.
	TStakeOrderModel interface {
		tStakeOrderModel
		FindPage(ctx context.Context, filter StakeOrderPageFilter, cursor int64, limit int64) ([]*TStakeOrder, int64, error)
		SumStakeAmountByStatuses(ctx context.Context, tenantID, user_id, productID int64, statuses []int64) (decimal.Decimal, error)
		ClaimOperation(ctx context.Context, id int64, operationNo string, now int64, allowedStatuses []int64) (bool, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TStakeOrder, error)
	}

	customTStakeOrderModel struct {
		*defaultTStakeOrderModel
	}
)

// NewTStakeOrderModel returns a model for the database table.
func NewTStakeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeOrderModel {
	return &customTStakeOrderModel{
		defaultTStakeOrderModel: newTStakeOrderModel(conn, c, opts...),
	}
}

func (m *customTStakeOrderModel) ClaimOperation(ctx context.Context, id int64, operationNo string, now int64, allowedStatuses []int64) (bool, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.And("id = ?", id)
	builder.And("active_operation_no = ''")
	builder.InInt64("status", allowedStatuses)
	query := fmt.Sprintf("UPDATE %s SET active_operation_no=?,redeem_apply_times=?,version=version+1,update_times=? WHERE %s", m.table, builder.Where())
	args := []any{operationNo, now, now}
	args = append(args, builder.Args()...)
	key := fmt.Sprintf("%s%v", cacheTStakeOrderIdPrefix, id)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, args...)
	}, key)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTStakeOrderModel) FindOneForUpdate(ctx context.Context, id int64) (*TStakeOrder, error) {
	var item TStakeOrder
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? FOR UPDATE", tStakeOrderRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTStakeOrderModel) FindPage(ctx context.Context, filter StakeOrderPageFilter, cursor int64, limit int64) ([]*TStakeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.TenantId > 0 {
		builder.And("tenant_id = ?", filter.TenantId)
	}
	if filter.UserId > 0 {
		builder.And("user_id = ?", filter.UserId)
	}
	if filter.ProductId > 0 {
		builder.And("product_id = ?", filter.ProductId)
	}
	builder.EqString("order_no", filter.OrderNo)
	if filter.ProductName != "" {
		builder.LikeString("product_name", filter.ProductName)
	}
	builder.EqString("coin_symbol", filter.CoinSymbol)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("redeem_type", filter.RedeemType)
	builder.EqInt64("source", filter.Source)
	builder.GteInt64("start_times", filter.StartBegin)
	builder.LteInt64("start_times", filter.StartEnd)
	builder.GteInt64("end_times", filter.EndBegin)
	builder.LteInt64("end_times", filter.EndEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customTStakeOrderModel) SumStakeAmountByStatuses(ctx context.Context, tenantID, user_id, productID int64, statuses []int64) (decimal.Decimal, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	builder.And("user_id = ?", user_id)
	builder.And("product_id = ?", productID)
	builder.InInt64("status", statuses)

	var total decimal.NullDecimal
	query := fmt.Sprintf("SELECT COALESCE(SUM(stake_amount), 0) FROM %s WHERE %s", m.table, builder.Where())
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, builder.Args()...); err != nil {
		return decimal.Zero, err
	}
	if !total.Valid {
		return decimal.Zero, nil
	}
	return total.Decimal, nil
}
