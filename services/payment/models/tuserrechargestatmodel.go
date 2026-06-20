package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TUserRechargeStatModel = (*customTUserRechargeStatModel)(nil)

type (
	UserRechargeStatPageFilter struct {
		TenantId              int64
		UserId                int64
		SuccessTotalAmountMin int64
		SuccessTotalAmountMax int64
	}

	// TUserRechargeStatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserRechargeStatModel.
	TUserRechargeStatModel interface {
		tUserRechargeStatModel
		FindPage(ctx context.Context, filter UserRechargeStatPageFilter, cursor int64, limit int64) ([]*TUserRechargeStat, int64, error)
	}

	customTUserRechargeStatModel struct {
		*defaultTUserRechargeStatModel
	}
)

// NewTUserRechargeStatModel returns a model for the database table.
func NewTUserRechargeStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserRechargeStatModel {
	return &customTUserRechargeStatModel{
		defaultTUserRechargeStatModel: newTUserRechargeStatModel(conn, c, opts...),
	}
}

func (m *defaultTUserRechargeStatModel) FindPage(ctx context.Context, filter UserRechargeStatPageFilter, cursor int64, limit int64) ([]*TUserRechargeStat, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.GteInt64("success_total_amount", filter.SuccessTotalAmountMin)
	builder.LteInt64("success_total_amount", filter.SuccessTotalAmountMax)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tUserRechargeStatRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TUserRechargeStat
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
