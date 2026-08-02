package models

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
	"wklive/proto/option"
)

var _ TOptionComboOrderModel = (*customTOptionComboOrderModel)(nil)

type (
	OptionComboOrderPageFilter struct {
		TenantId         int64
		UserId           int64
		AccountId        int64
		ComboNo          string
		UnderlyingSymbol string
		Status           int64
		CreateTimeStart  int64
		CreateTimeEnd    int64
	}

	// TOptionComboOrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionComboOrderModel.
	TOptionComboOrderModel interface {
		tOptionComboOrderModel
		FindOneByTenantIdUserIdClientComboIdNoCache(ctx context.Context, tenantId int64, userId int64, clientComboId string) (*TOptionComboOrder, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionComboOrder, error)
		FindPage(ctx context.Context, filter OptionComboOrderPageFilter, cursor, limit int64) ([]*TOptionComboOrder, int64, error)
		FindMatchCandidates(ctx context.Context, tenantId int64, strategyKey string, excludeUserId int64, minNetPrice decimal.Decimal, limit int64) ([]*TOptionComboOrder, error)
	}

	customTOptionComboOrderModel struct {
		*defaultTOptionComboOrderModel
	}
)

// NewTOptionComboOrderModel returns a model for the database table.
func NewTOptionComboOrderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionComboOrderModel {
	return &customTOptionComboOrderModel{
		defaultTOptionComboOrderModel: newTOptionComboOrderModel(conn, c, opts...),
	}
}

// FindOneByTenantIdUserIdClientComboIdNoCache is the authoritative lookup for
// idempotent create/replay. A cached miss must never hide a row committed by a
// concurrent winner of uk_option_combo_client.
func (m *defaultTOptionComboOrderModel) FindOneByTenantIdUserIdClientComboIdNoCache(
	ctx context.Context, tenantId int64, userId int64, clientComboId string,
) (*TOptionComboOrder, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id=? AND user_id=? AND client_combo_id=? LIMIT 1",
		tOptionComboOrderRows, m.table,
	)
	var item TOptionComboOrder
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, userId, clientComboId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionComboOrderModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionComboOrder, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE", tOptionComboOrderRows, m.table)
	var item TOptionComboOrder
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionComboOrderModel) FindPage(
	ctx context.Context, filter OptionComboOrderPageFilter, cursor, limit int64,
) ([]*TOptionComboOrder, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.LikeString("combo_no", filter.ComboNo)
	builder.LikeString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqInt64("status", filter.Status)
	builder.GteInt64("create_times", filter.CreateTimeStart)
	builder.LteInt64("create_times", filter.CreateTimeEnd)
	where := builder.Where()
	args := builder.Args()

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	listQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionComboOrderRows, m.table, where)
	if cursor > 0 {
		listQuery += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listQuery += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var list []*TOptionComboOrder
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTOptionComboOrderModel) FindMatchCandidates(
	ctx context.Context,
	tenantId int64,
	strategyKey string,
	excludeUserId int64,
	minNetPrice decimal.Decimal,
	limit int64,
) ([]*TOptionComboOrder, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND strategy_key=? AND user_id<>?
  AND status IN (?,?) AND unfilled_qty>0 AND net_price>=?
ORDER BY net_price DESC, id ASC
LIMIT ?`, tOptionComboOrderRows, m.table)
	var list []*TOptionComboOrder
	if err := m.QueryRowsNoCacheCtx(
		ctx, &list, query,
		tenantId, strategyKey, excludeUserId,
		int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE),
		int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED),
		minNetPrice, limit,
	); err != nil {
		return nil, err
	}
	return list, nil
}
