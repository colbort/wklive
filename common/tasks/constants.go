package tasks

import (
	"context"
	"fmt"
	"time"
	bus "wklive/common/bus/redis"
)

const (
	channel = "system:scheduled-tasks"

	ServiceItick   = "itick"
	ServiceOption  = "option"
	ServiceStaking = "staking"
	ServiceTrade   = "trade"

	ActionItickSyncProducts = "SyncProducts"
	ActionItickSyncKlines   = "SyncKlines"

	ActionOptionProcessContractLifecycle = "ProcessContractLifecycle"
	ActionOptionCleanMarketSnapshots     = "CleanMarketSnapshots"

	ActionStakingProcessRewardsAndSettleOrders = "ProcessRewardsAndSettleOrders"

	ActionTradeProcessOrderMatching       = "ProcessOrderMatching"
	ActionTradeProcessPositions           = "ProcessPositions"
	ActionTradeProcessContractSettlements = "ProcessContractSettlements"
	ActionTradeProcessTradeEvents         = "ProcessTradeEvents"
	ActionTradeExpireRiskLimits           = "ExpireRiskLimits"
)

type Message struct {
	ID        string `json:"id"`
	Service   string `json:"service"`
	Action    string `json:"action"`
	TenantID  int64  `json:"tenantId,omitempty"`
	JobID     int64  `json:"jobId,omitempty"`
	JobName   string `json:"jobName,omitempty"`
	CreatedAt int64  `json:"createdAt"`
}

type PublishOptions struct {
	TenantID int64
	JobID    int64
	JobName  string
}

type Handler func(context.Context, Message) error

func NewMessage(service, action string, tenantID int64, jobID int64, jobName string) Message {
	now := time.Now().UnixMilli()
	return Message{
		ID:        fmt.Sprintf("%d", now),
		Service:   service,
		Action:    action,
		TenantID:  tenantID,
		JobID:     jobID,
		JobName:   jobName,
		CreatedAt: now,
	}
}

func Publish(ctx context.Context, publisher *bus.Publisher, service string, action string, opts PublishOptions) error {
	if publisher == nil {
		return fmt.Errorf("task publisher is nil")
	}

	msg := NewMessage(service, action, opts.TenantID, opts.JobID, opts.JobName)
	return publisher.Publish(ctx, channel, msg)
}

func SubscribeService(ctx context.Context, subscriber *bus.Subscriber, service string, handler Handler) error {
	if subscriber == nil {
		return fmt.Errorf("task subscriber is nil")
	}
	if service == "" {
		return fmt.Errorf("task service is empty")
	}
	if handler == nil {
		return fmt.Errorf("task handler is nil")
	}

	return subscriber.Subscribe(ctx, channel, func(ctx context.Context, msg bus.Message) error {
		var payload Message
		if err := bus.Decode(msg, &payload); err != nil {
			return err
		}
		if payload.Service != service {
			return nil
		}
		return handler(ctx, payload)
	})
}
