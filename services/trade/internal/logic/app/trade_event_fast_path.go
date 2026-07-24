package applogic

import (
	"context"
	"sync"

	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	tradeEventFastPathWorkers = 4
	tradeEventFastPathSize    = 1024
)

type tradeEventFastPathJob struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	event  realtime.Event
}

var (
	tradeEventFastPathOnce  sync.Once
	tradeEventFastPathQueue = make(chan tradeEventFastPathJob, tradeEventFastPathSize)
)

func enqueueTradeEventFastPath(ctx context.Context, svcCtx *svc.ServiceContext, event realtime.Event) {
	tradeEventFastPathOnce.Do(func() {
		for range tradeEventFastPathWorkers {
			go func() {
				for job := range tradeEventFastPathQueue {
					if err := publishTradeOutboxEvent(job.ctx, job.svcCtx, job.event); err != nil {
						// The durable outbox remains pending and ProcessTradeEvents
						// will retry it. The fast path must not fail the order.
						logx.WithContext(job.ctx).Errorf(
							"publish trade event fast path failed, eventNo=%s err=%v",
							job.event.EventNo, err,
						)
					}
				}
			}()
		}
	})

	job := tradeEventFastPathJob{ctx: context.WithoutCancel(ctx), svcCtx: svcCtx, event: event}
	select {
	case tradeEventFastPathQueue <- job:
	default:
		// Backpressure must not create unbounded goroutines or delay PlaceOrder.
		// The event is already durable and will be picked up by the outbox task.
		logx.WithContext(ctx).Errorf(
			"trade event fast path queue full, deferred to outbox dispatcher eventNo=%s",
			event.EventNo,
		)
	}
}
