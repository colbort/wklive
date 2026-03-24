package bootstrap

import (
	"context"
	"wklive/services/system/internal/svc"
)

func LoadJobs(ctx *svc.ServiceContext) error {
	jobs, err := ctx.JobModel.FindEnabledJobs(context.Background())
	if err != nil {
		return err
	}
	return ctx.Cron.LoadJobs(jobs)
}
