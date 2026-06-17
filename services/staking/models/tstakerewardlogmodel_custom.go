package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeRewardLogPageFilter struct {
	TenantId     int64
	UserId       int64
	OrderId      int64
	ProductId    int64
	OrderNo      string
	RewardType   int64
	RewardStatus int64
	RewardBegin  int64
	RewardEnd    int64
}

type StakeRewardLogModel interface {
	tStakeRewardLogModel
	FindPage(ctx context.Context, filter StakeRewardLogPageFilter, cursor int64, limit int64) ([]*TStakeRewardLog, int64, error)
}

func (m *defaultTStakeRewardLogModel) FindPage(ctx context.Context, filter StakeRewardLogPageFilter, cursor int64, limit int64) ([]*TStakeRewardLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	if filter.TenantId > 0 {
		builder.And("tenant_id = ?", filter.TenantId)
	}
	if filter.UserId > 0 {
		builder.And("user_id = ?", filter.UserId)
	}
	if filter.OrderId > 0 {
		builder.And("order_id = ?", filter.OrderId)
	}
	if filter.ProductId > 0 {
		builder.And("product_id = ?", filter.ProductId)
	}
	builder.EqString("order_no", filter.OrderNo)
	builder.EqInt64("reward_type", filter.RewardType)
	builder.EqInt64("reward_status", filter.RewardStatus)
	builder.GteInt64("reward_times", filter.RewardBegin)
	builder.LteInt64("reward_times", filter.RewardEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tStakeRewardLogRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TStakeRewardLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
