package models

import "context"

type UserRechargeStatModel interface {
	tUserRechargeStatModel
	FindPage(ctx context.Context, tenantId int64, userId int64, successTotalAmountMin int64, successTotalAmountMax int64, cursor int64, limit int64) ([]*TUserRechargeStat, int64, error)
}
