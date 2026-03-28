package tasks

import (
	"context"
	"fmt"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("orderJob.autoCancel", "Auto Cancel Order", autoCancel)
}

func autoCancel(ctx context.Context, job *models.SysJob) error {
	fmt.Println("执行自动取消订单任务")
	return nil
}
