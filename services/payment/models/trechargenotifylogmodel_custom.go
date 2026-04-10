package models

import "context"

type RechargeNotifyLogModel interface {
	tRechargeNotifyLogModel
	FindPage(ctx context.Context, tenantId int64, orderNo string, orderId int64, platformId int64, channelId int64, notifyStatus int64, signResult int64, createTimeStart int64, createTimeEnd int64, cursor int64, limit int64) ([]*TRechargeNotifyLog, int64, error)
	FindOneByOrderId(ctx context.Context, orderId int64) (*TRechargeNotifyLog, error)
}
