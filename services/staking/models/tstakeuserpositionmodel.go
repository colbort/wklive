package models

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeUserPositionModel = (*customTStakeUserPositionModel)(nil)

type (
	// TStakeUserPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeUserPositionModel.
	TStakeUserPositionModel interface {
		tStakeUserPositionModel
		Ensure(ctx context.Context, tenantId, userId, productId, now int64) error
		ReserveAmount(ctx context.Context, tenantId, userId, productId int64, amount, userLimit decimal.Decimal, now int64) (bool, error)
		ReleaseAmount(ctx context.Context, tenantId, userId, productId int64, amount decimal.Decimal, now int64) error
	}

	customTStakeUserPositionModel struct {
		*defaultTStakeUserPositionModel
	}
)

// NewTStakeUserPositionModel returns a model for the database table.
func NewTStakeUserPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeUserPositionModel {
	return &customTStakeUserPositionModel{
		defaultTStakeUserPositionModel: newTStakeUserPositionModel(conn, c, opts...),
	}
}

func (m *defaultTStakeUserPositionModel) Ensure(ctx context.Context, tenantId, userId, productId, now int64) error {
	_, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_stake_user_position
		(tenant_id,user_id,product_id,staked_amount,version,create_times,update_times)
		VALUES (?,?,?,0,1,?,?) ON DUPLICATE KEY UPDATE id=id`, tenantId, userId, productId, now, now)
	return err
}

func (m *defaultTStakeUserPositionModel) ReserveAmount(ctx context.Context, tenantId, userId, productId int64, amount, userLimit decimal.Decimal, now int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `UPDATE t_stake_user_position
		SET staked_amount=staked_amount+?,version=version+1,update_times=?
		WHERE tenant_id=? AND user_id=? AND product_id=?
		  AND (?=0 OR staked_amount+?<=?)`, amount, now, tenantId, userId, productId, userLimit, amount, userLimit)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *defaultTStakeUserPositionModel) ReleaseAmount(ctx context.Context, tenantId, userId, productId int64, amount decimal.Decimal, now int64) error {
	_, err := m.ExecNoCacheCtx(ctx, `UPDATE t_stake_user_position
		SET staked_amount=GREATEST(staked_amount-?,0),version=version+1,update_times=?
		WHERE tenant_id=? AND user_id=? AND product_id=?`, amount, now, tenantId, userId, productId)
	return err
}
