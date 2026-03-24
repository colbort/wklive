package tasks

import (
	"context"
	"fmt"
	"time"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("reportJob.dailyReport", dailyReport)
}

func dailyReport(ctx context.Context, job *models.SysJob) error {
	fmt.Println("开始执行日报任务:", job.JobName)

	select {
	case <-time.After(5 * time.Second):
		fmt.Println("日报任务执行完成")
		return nil
	case <-ctx.Done():
		fmt.Println("日报任务被取消")
		return ctx.Err()
	}
}
