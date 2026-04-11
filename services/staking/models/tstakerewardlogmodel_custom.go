package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type StakeRewardLogModel interface {
	tStakeRewardLogModel
	FindPage(ctx context.Context, tenantID int64, cursor, limit int64, uid, orderID, productID int64, orderNo string, rewardType, rewardStatus int64, rewardBegin, rewardEnd int64) ([]*TStakeRewardLog, int64, error)
}

func (m *defaultTStakeRewardLogModel) FindPage(ctx context.Context, tenantID int64, cursor, limit int64, uid, orderID, productID int64, orderNo string, rewardType, rewardStatus int64, rewardBegin, rewardEnd int64) ([]*TStakeRewardLog, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.And("tenant_id = ?", tenantID)
	if uid > 0 {
		builder.And("uid = ?", uid)
	}
	if orderID > 0 {
		builder.And("order_id = ?", orderID)
	}
	if productID > 0 {
		builder.And("product_id = ?", productID)
	}
	builder.EqString("order_no", orderNo)
	builder.EqInt64("reward_type", rewardType)
	builder.EqInt64("reward_status", rewardStatus)
	builder.GteInt64("reward_times", rewardBegin)
	builder.LteInt64("reward_times", rewardEnd)

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
