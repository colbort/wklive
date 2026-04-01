package models

import "context"

type RechargeOrderModel interface {
	tRechargeOrderModel
	FindPage(ctx context.Context, tenantId int64, userId int64, orderNo string, status int64, cursor int64, limit int64) ([]*TRechargeOrder, int64, error)
}
