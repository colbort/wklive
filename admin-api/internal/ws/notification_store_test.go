package ws

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"wklive/common/notify"
	"wklive/proto/system"

	"google.golang.org/grpc"
)

type notificationSystemClientStub struct {
	recordRequest  *system.RecordAdminNotificationReq
	ackRequest     *system.AcknowledgeAdminNotificationReq
	claimResponse  *system.ClaimDueAdminNotificationsResp
	releaseRequest *system.ReleaseAdminNotificationEscalationReq
}

func (s *notificationSystemClientStub) RecordAdminNotification(
	_ context.Context,
	request *system.RecordAdminNotificationReq,
	_ ...grpc.CallOption,
) (*system.AdminNotificationIncidentResp, error) {
	s.recordRequest = request
	return &system.AdminNotificationIncidentResp{
		Data: &system.AdminNotificationIncident{Id: 1},
	}, nil
}

func (s *notificationSystemClientStub) AcknowledgeAdminNotification(
	_ context.Context,
	request *system.AcknowledgeAdminNotificationReq,
	_ ...grpc.CallOption,
) (*system.AdminNotificationIncidentResp, error) {
	s.ackRequest = request
	return &system.AdminNotificationIncidentResp{
		Data: &system.AdminNotificationIncident{
			AcknowledgedAt: request.AcknowledgedAt,
			AcknowledgedBy: request.UserId,
		},
	}, nil
}

func (s *notificationSystemClientStub) ClaimDueAdminNotifications(
	_ context.Context,
	_ *system.ClaimDueAdminNotificationsReq,
	_ ...grpc.CallOption,
) (*system.ClaimDueAdminNotificationsResp, error) {
	if s.claimResponse == nil {
		return &system.ClaimDueAdminNotificationsResp{}, nil
	}
	return s.claimResponse, nil
}

func (s *notificationSystemClientStub) ReleaseAdminNotificationEscalation(
	_ context.Context,
	request *system.ReleaseAdminNotificationEscalationReq,
	_ ...grpc.CallOption,
) (*system.RespBase, error) {
	s.releaseRequest = request
	return &system.RespBase{}, nil
}

type notificationPublisherStub struct {
	event notify.Event
	err   error
}

func (p *notificationPublisherStub) Publish(
	_ context.Context,
	_ string,
	value any,
) error {
	p.event, _ = value.(notify.Event)
	return p.err
}

func TestNotificationStoreRecordsOperationalEvent(t *testing.T) {
	client := &notificationSystemClientStub{}
	store := NewNotificationStore(
		client,
		nil,
		NotificationStoreConfig{AckTimeout: 7 * time.Minute},
	)
	event := notify.Event{
		ID:        "event-1",
		Type:      "snapshot_outbox",
		Level:     "error",
		Title:     "outbox unhealthy",
		Message:   "oldest pending exceeded limit",
		Source:    "market",
		TenantID:  7,
		BizNo:     "snapshot-outbox",
		CreatedAt: 1234,
		Data: map[string]any{
			"state":    "firing",
			"severity": "critical",
		},
	}
	if err := store.RecordEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if client.recordRequest == nil ||
		client.recordRequest.AlertKey != "snapshot-outbox" ||
		client.recordRequest.AckTimeoutMs != int64((7*time.Minute)/time.Millisecond) {
		t.Fatalf("request=%+v", client.recordRequest)
	}
}

func TestNotificationStoreAcknowledgesOnlyAuthorizedTenant(t *testing.T) {
	client := &notificationSystemClientStub{}
	store := NewNotificationStore(client, nil, NotificationStoreConfig{})
	connection := NewConnection(
		nil,
		nil,
		10,
		"operator",
		7,
		false,
		[]string{"market:snapshot-outbox:list"},
		store,
	)
	payload, _ := json.Marshal(clientMessage{
		Action:    "ack",
		EventType: "snapshot_outbox",
		AlertKey:  "snapshot-outbox",
		TenantID:  8,
		Reason:    "manual acknowledgement",
	})
	var response clientMessageResponse
	_ = json.Unmarshal(store.HandleClientMessage(context.Background(), connection, payload), &response)
	if response.OK || client.ackRequest != nil {
		t.Fatalf("cross-tenant response=%+v request=%+v", response, client.ackRequest)
	}

	payload, _ = json.Marshal(clientMessage{
		Action:    "ack",
		EventType: "snapshot_outbox",
		AlertKey:  "snapshot-outbox",
		TenantID:  7,
		Reason:    "manual acknowledgement",
	})
	_ = json.Unmarshal(store.HandleClientMessage(context.Background(), connection, payload), &response)
	if !response.OK || client.ackRequest == nil || client.ackRequest.UserId != 10 {
		t.Fatalf("response=%+v request=%+v", response, client.ackRequest)
	}
}

func TestNotificationStoreEscalatesAndReleasesFailedPublish(t *testing.T) {
	payload, _ := json.Marshal(notify.Event{
		ID:        "event-1",
		Type:      "price_engine_input",
		Level:     "error",
		Title:     "price missing",
		Message:   "no source",
		Source:    "market",
		BizNo:     "formula-1",
		CreatedAt: 1000,
		Data:      map[string]any{"state": "firing"},
	})
	client := &notificationSystemClientStub{
		claimResponse: &system.ClaimDueAdminNotificationsResp{
			Data: []*system.AdminNotificationIncident{{
				Id:              11,
				EventType:       "price_engine_input",
				AlertKey:        "formula-1",
				Title:           "price missing",
				PayloadJson:     string(payload),
				EscalationLevel: 2,
				Version:         3,
			}},
		},
	}
	publisher := &notificationPublisherStub{}
	store := NewNotificationStore(client, publisher, NotificationStoreConfig{})
	if err := store.EscalateOnce(context.Background(), 2000); err != nil {
		t.Fatal(err)
	}
	if publisher.event.Data["escalationLevel"] != int64(2) &&
		publisher.event.Data["escalationLevel"] != float64(2) {
		t.Fatalf("event=%+v", publisher.event)
	}
	if client.releaseRequest != nil {
		t.Fatalf("unexpected release=%+v", client.releaseRequest)
	}

	publisher.err = errors.New("kafka unavailable")
	if err := store.EscalateOnce(context.Background(), 3000); err == nil {
		t.Fatal("expected publish error")
	}
	if client.releaseRequest == nil ||
		client.releaseRequest.Id != 11 ||
		client.releaseRequest.ClaimedLevel != 2 {
		t.Fatalf("release=%+v", client.releaseRequest)
	}
}
