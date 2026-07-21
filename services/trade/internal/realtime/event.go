package realtime

import (
	"context"
	"wklive/common/mq/kafka"
)

const Channel = "trade:domain-events"

const (
	ConsumerTradeRealtime = "trade-realtime"
	PayloadVersionV1      = int64(1)
	ClaimLeaseMillis      = int64(30_000)
)

const (
	EventOrderAccepted = "ORDER_ACCEPTED"
	EventFillCreated   = "FILL_CREATED"
	EventPositionFill  = "POSITION_FILL_REQUIRED"
)

type Event struct {
	Version    int64  `json:"version"`
	Consumer   string `json:"consumer"`
	ClaimToken string `json:"claim_token"`
	EventNo    string `json:"event_no"`
	Type       string `json:"type"`
	TenantID   int64  `json:"tenant_id"`
	BizID      string `json:"biz_id"`
	OrderID    int64  `json:"order_id,omitempty"`
	FillID     int64  `json:"fill_id,omitempty"`
	Payload    string `json:"payload,omitempty"`
}

func Publish(ctx context.Context, publisher *mq.Publisher, event Event) error {
	return publisher.Publish(ctx, Channel, event)
}
