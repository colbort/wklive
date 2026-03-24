package tasks

import (
	"context"
	"fmt"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("demoJob.test", testJob)
}

func testJob(ctx context.Context, job *models.SysJob) error {
	fmt.Println("执行 demoJob.test, job:", job.JobName)
	return nil
}
