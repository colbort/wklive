package tasks

import (
	"context"
	"fmt"
	"sync"

	bus "wklive/common/bus/redis"
	"wklive/common/tasks"
	"wklive/services/system/models"
)

var (
	taskPublisherMu sync.RWMutex
	taskPublisher   *bus.Publisher
)

func InitTaskPublisher(publisher *bus.Publisher) {
	taskPublisherMu.Lock()
	defer taskPublisherMu.Unlock()
	taskPublisher = publisher
}

func publishTask(ctx context.Context, job *models.SysJob, service, action string) error {
	var jobID int64
	var jobName string
	if job != nil {
		jobID = job.Id
		jobName = job.JobName
	}

	taskPublisherMu.RLock()
	publisher := taskPublisher
	taskPublisherMu.RUnlock()
	if publisher == nil {
		return fmt.Errorf("task publisher is not initialized")
	}

	return tasks.Publish(ctx, publisher, service, action, tasks.PublishOptions{
		JobID:   jobID,
		JobName: jobName,
	})
}
