package models

import "context"

type WithdrawNotifyLogModel interface {
	tWithdrawNotifyLogModel
	FindPage(ctx context.Context, tenantId int64, orderNo string, notifyStatus int64, cursor int64, limit int64) ([]*TWithdrawNotifyLog, int64, error)
}
