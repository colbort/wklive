package plugins

import (
	"context"
	"fmt"
	"time"
	"wklive/services/system/models"
)

func RegisterDefaultHandlers(mgr *CronManager) {
	mgr.RegisterHandler("demoJob.test", func(ctx context.Context, job *models.SysJob) error {
		fmt.Println("执行 demoJob.test, job:", job.JobName)
		return nil
	})

	mgr.RegisterHandler("reportJob.dailyReport", func(ctx context.Context, job *models.SysJob) error {
		fmt.Println("开始执行日报任务:", job.JobName)
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("日报任务执行完成")
			return nil
		case <-ctx.Done():
			fmt.Println("日报任务被取消")
			return ctx.Err()
		}
	})

	mgr.RegisterHandler("orderJob.autoCancel", func(ctx context.Context, job *models.SysJob) error {
		fmt.Println("执行自动取消订单任务")
		return nil
	})
}
