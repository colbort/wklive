package bootstrap

import (
	"context"
	"fmt"
	"time"

	"wklive/services/system/internal/svc"
)

func LoadJobs(ctx *svc.ServiceContext) error {
	jobs, err := ctx.JobModel.FindEnabledJobs(context.Background())
	if err != nil {
		return err
	}
	return ctx.Cron.LoadJobs(jobs)
}

// LoadJobsWithRetry prevents the service from becoming healthy without any
// scheduled jobs when MySQL is briefly unavailable during container startup.
func LoadJobsWithRetry(ctx *svc.ServiceContext, attempts int, delay time.Duration) error {
	return retry(attempts, delay, func() error { return LoadJobs(ctx) })
}

func retry(attempts int, delay time.Duration, run func() error) error {
	if attempts <= 0 {
		return fmt.Errorf("attempts must be positive")
	}
	if run == nil {
		return fmt.Errorf("retry function is nil")
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if lastErr = run(); lastErr == nil {
			return nil
		}
		if attempt < attempts && delay > 0 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}
