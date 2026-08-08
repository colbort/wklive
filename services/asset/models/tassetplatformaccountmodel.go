package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TAssetPlatformAccountModel = (*customTAssetPlatformAccountModel)(nil)

type (
	// TAssetPlatformAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetPlatformAccountModel.
	TAssetPlatformAccountModel interface {
		tAssetPlatformAccountModel
		FindOneForUpdate(ctx context.Context, tenantID int64, accountType, coin string) (*TAssetPlatformAccount, error)
		AddAvailable(ctx context.Context, id int64, amount decimal.Decimal, now int64) error
		SubAvailable(ctx context.Context, id int64, amount decimal.Decimal, now int64) (bool, error)
		SubAvailableWithFloor(ctx context.Context, id int64, amount, floor decimal.Decimal, now int64) (bool, error)
	}

	customTAssetPlatformAccountModel struct {
		*defaultTAssetPlatformAccountModel
	}
)

// NewTAssetPlatformAccountModel returns a model for the database table.
func NewTAssetPlatformAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetPlatformAccountModel {
	return &customTAssetPlatformAccountModel{
		defaultTAssetPlatformAccountModel: newTAssetPlatformAccountModel(conn, c, opts...),
	}
}

func (m *customTAssetPlatformAccountModel) SubAvailableWithFloor(
	ctx context.Context,
	id int64,
	amount decimal.Decimal,
	floor decimal.Decimal,
	now int64,
) (bool, error) {
	item, err := m.findOneNoCache(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTAssetPlatformAccountIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v", cacheTAssetPlatformAccountTenantIdAccountTypeCoinPrefix,
		item.TenantId, item.AccountType, item.Coin)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_asset_platform_account
SET available_amount=available_amount-?,version=version+1,update_times=?
WHERE id=? AND status=1 AND account_type='OPTION_BACKSTOP'
  AND available_amount-?>=?`, amount, now, id, amount, floor)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTAssetPlatformAccountModel) FindOneForUpdate(ctx context.Context, tenantID int64, accountType, coin string) (*TAssetPlatformAccount, error) {
	var row TAssetPlatformAccount
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND account_type=? AND coin=? AND status=1 FOR UPDATE", tAssetPlatformAccountRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, tenantID, accountType, coin); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *customTAssetPlatformAccountModel) AddAvailable(ctx context.Context, id int64, amount decimal.Decimal, now int64) error {
	item, err := m.findOneNoCache(ctx, id)
	if err != nil {
		return err
	}
	idKey := fmt.Sprintf("%s%v", cacheTAssetPlatformAccountIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v", cacheTAssetPlatformAccountTenantIdAccountTypeCoinPrefix, item.TenantId, item.AccountType, item.Coin)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, "UPDATE t_asset_platform_account SET available_amount=available_amount+?,version=version+1,update_times=? WHERE id=? AND status=1", amount, now, id)
	}, idKey, uniqueKey)
	return err
}

func (m *customTAssetPlatformAccountModel) SubAvailable(ctx context.Context, id int64, amount decimal.Decimal, now int64) (bool, error) {
	item, err := m.findOneNoCache(ctx, id)
	if err != nil {
		return false, err
	}
	idKey := fmt.Sprintf("%s%v", cacheTAssetPlatformAccountIdPrefix, id)
	uniqueKey := fmt.Sprintf("%s%v:%v:%v", cacheTAssetPlatformAccountTenantIdAccountTypeCoinPrefix, item.TenantId, item.AccountType, item.Coin)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, "UPDATE t_asset_platform_account SET available_amount=available_amount-?,version=version+1,update_times=? WHERE id=? AND status=1 AND available_amount>=?", amount, now, id, amount)
	}, idKey, uniqueKey)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *customTAssetPlatformAccountModel) findOneNoCache(ctx context.Context, id int64) (*TAssetPlatformAccount, error) {
	var item TAssetPlatformAccount
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1", tAssetPlatformAccountRows, m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}
