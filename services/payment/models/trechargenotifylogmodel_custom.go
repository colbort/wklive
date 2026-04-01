package models

import "context"

type RechargeNotifyLogModel interface {
	tRechargeNotifyLogModel
	FindPage(ctx context.Context, tenantId int64, orderNo string, notifyStatus int64, cursor int64, limit int64) ([]*TRechargeNotifyLog, int64, error)
}
