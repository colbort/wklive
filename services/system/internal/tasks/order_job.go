package tasks

import (
	"context"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func init() {
	cronx.Register("orderJob.autoCancel", "Auto Cancel Order", autoCancel)
}

func autoCancel(ctx context.Context, job *models.SysJob) error {
	logx.Info("执行自动取消订单任务")
	return nil
}
