package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TStakeProductModel = (*customTStakeProductModel)(nil)

type (
	StakeProductPageFilter struct {
		TenantId    int64
		ProductNo   string
		ProductName string
		CoinSymbol  string
		ProductType int64
		Status      int64
	}

	// TStakeProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeProductModel.
	TStakeProductModel interface {
		tStakeProductModel
		FindPage(ctx context.Context, filter StakeProductPageFilter, cursor int64, limit int64) ([]*TStakeProduct, int64, error)
		ReserveStakeAmount(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error)
		ReleaseStakeAmount(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) error
	}

	customTStakeProductModel struct {
		*defaultTStakeProductModel
	}
)

// NewTStakeProductModel returns a model for the database table.
func NewTStakeProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeProductModel {
	return &customTStakeProductModel{
		defaultTStakeProductModel: newTStakeProductModel(conn, c, opts...),
	}
}

func (m *defaultTStakeProductModel) ReserveStakeAmount(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) (bool, error) {
	key := fmt.Sprintf("%s%v", cacheTStakeProductIdPrefix, id)
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_stake_product
			SET staked_amount=staked_amount+?,update_times=?
			WHERE id=? AND status=2
			  AND (total_amount=0 OR staked_amount+?<=total_amount)`, amount, updateTimes, id, amount)
	}, key)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func (m *defaultTStakeProductModel) ReleaseStakeAmount(ctx context.Context, id int64, amount decimal.Decimal, updateTimes int64) error {
	key := fmt.Sprintf("%s%v", cacheTStakeProductIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, `UPDATE t_stake_product
			SET staked_amount=GREATEST(staked_amount-?,0),update_times=? WHERE id=?`, amount, updateTimes, id)
	}, key)
	return err
}

func (m *defaultTStakeProductModel) FindPage(ctx context.Context, filter StakeProductPageFilter, cursor int64, limit int64) ([]*TStakeProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.TenantId > 0 {
		builder.And("tenant_id = ?", filter.TenantId)
	}
	builder.EqString("product_no", filter.ProductNo)
	if filter.ProductName != "" {
		builder.LikeString("product_name", filter.ProductName)
	}
	builder.EqString("coin_symbol", filter.CoinSymbol)
	builder.EqInt64("product_type", filter.ProductType)
	builder.EqInt64("status", filter.Status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeProductRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY sort DESC, id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
