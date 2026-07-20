package realtime

import (
	"context"

	bus "wklive/common/bus/redis"
)

const Channel = "trade:domain-events"

const (
	EventOrderAccepted = "ORDER_ACCEPTED"
	EventFillCreated   = "FILL_CREATED"
)

type Event struct {
	EventNo  string `json:"event_no"`
	Type     string `json:"type"`
	TenantID int64  `json:"tenant_id"`
	BizID    string `json:"biz_id"`
	OrderID  int64  `json:"order_id,omitempty"`
	FillID   int64  `json:"fill_id,omitempty"`
}

func Publish(ctx context.Context, publisher *bus.Publisher, event Event) error {
	return publisher.Publish(ctx, Channel, event)
}
