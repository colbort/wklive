package models

import "context"

type WithdrawOrderModel interface {
	tWithdrawOrderModel
	FindPage(ctx context.Context, tenantId int64, userId int64, orderNo string, status int64, cursor int64, limit int64) ([]*TWithdrawOrder, int64, error)
}
