package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/common/sqlutil"
	"wklive/proto/payment"
)

var _ TRechargeOrderModel = (*customTRechargeOrderModel)(nil)

type (
	RechargeOrderPageFilter struct {
		TenantId     int64
		UserId       int64
		OrderNo      string
		Status       int64
		RechargeType int64
	}

	RechargeOrderCreditIdentity struct {
		TenantId   int64          `db:"tenant_id"`
		OrderNo    string         `db:"order_no"`
		BizOrderNo sql.NullString `db:"biz_order_no"`
	}

	// TRechargeOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRechargeOrderModel.
	TRechargeOrderModel interface {
		tRechargeOrderModel
		FindPage(ctx context.Context, filter RechargeOrderPageFilter, cursor int64, limit int64) ([]*TRechargeOrder, int64, error)
		FindCreditIdentityForUpdate(ctx context.Context, id int64) (*RechargeOrderCreditIdentity, error)
		MarkCreditSuccess(ctx context.Context, id int64, identity *RechargeOrderCreditIdentity, now int64) (bool, error)
		MarkCreditRetrying(ctx context.Context, id int64, identity *RechargeOrderCreditIdentity, retryCount, now int64, message string) (bool, error)
	}

	customTRechargeOrderModel struct {
		*defaultTRechargeOrderModel
	}
)

// NewTRechargeOrderModel returns a model for the database table.
func NewTRechargeOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRechargeOrderModel {
	return &customTRechargeOrderModel{
		defaultTRechargeOrderModel: newTRechargeOrderModel(conn, c, opts...),
	}
}

func (m *customTRechargeOrderModel) FindCreditIdentityForUpdate(ctx context.Context, id int64) (*RechargeOrderCreditIdentity, error) {
	var row RechargeOrderCreditIdentity
	query := fmt.Sprintf("SELECT tenant_id, order_no, biz_order_no FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTRechargeOrderModel) MarkCreditSuccess(ctx context.Context, id int64, identity *RechargeOrderCreditIdentity, now int64) (bool, error) {
	return m.updateCreditState(
		ctx,
		id,
		identity,
		"credit_status = ?, credited_time = ?, last_credit_error = '', update_times = ?",
		int64(payment.CreditStatus_CREDIT_STATUS_SUCCESS),
		now,
		now,
	)
}

func (m *customTRechargeOrderModel) MarkCreditRetrying(
	ctx context.Context,
	id int64,
	identity *RechargeOrderCreditIdentity,
	retryCount, now int64,
	message string,
) (bool, error) {
	return m.updateCreditState(
		ctx,
		id,
		identity,
		"credit_status = ?, credit_retry_count = ?, last_credit_error = ?, update_times = ?",
		int64(payment.CreditStatus_CREDIT_STATUS_PROCESSING),
		retryCount,
		message,
		now,
	)
}

func (m *customTRechargeOrderModel) updateCreditState(
	ctx context.Context,
	id int64,
	identity *RechargeOrderCreditIdentity,
	setClause string,
	setArgs ...any,
) (bool, error) {
	if identity == nil {
		return false, nil
	}
	idKey := fmt.Sprintf("%s%v", cacheTRechargeOrderIdPrefix, id)
	orderNoKey := fmt.Sprintf("%s%v", cacheTRechargeOrderOrderNoPrefix, identity.OrderNo)
	bizOrderNoKey := fmt.Sprintf("%s%v:%v", cacheTRechargeOrderTenantIdBizOrderNoPrefix, identity.TenantId, identity.BizOrderNo)
	args := append(append([]any{}, setArgs...), id, identity.TenantId, identity.OrderNo)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ? AND tenant_id = ? AND order_no = ?", m.table, setClause)
		return conn.ExecCtx(ctx, query, args...)
	}, idKey, orderNoKey, bizOrderNoKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTRechargeOrderModel) FindPage(ctx context.Context, filter RechargeOrderPageFilter, cursor int64, limit int64) ([]*TRechargeOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("status", filter.Status)
	builder.EqInt64("recharge_type", filter.RechargeType)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tRechargeOrderRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TRechargeOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
