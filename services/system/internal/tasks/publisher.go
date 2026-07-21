package tasks

import (
	"context"
	"fmt"
	"sync"

	"wklive/common/mq/kafka"
	"wklive/common/tasks"
	"wklive/services/system/models"
)

var (
	taskPublisherMu sync.RWMutex
	taskPublisher   *mq.Publisher
)

func InitTaskPublisher(publisher *mq.Publisher) {
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
