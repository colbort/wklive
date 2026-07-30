package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mq "wklive/common/mq/kafka"
	"wklive/common/notify"
	"wklive/proto/system"

	"google.golang.org/grpc"
)

const escalationRetryDelay = time.Minute

type (
	NotificationSystemClient interface {
		RecordAdminNotification(
			context.Context,
			*system.RecordAdminNotificationReq,
			...grpc.CallOption,
		) (*system.AdminNotificationIncidentResp, error)
		AcknowledgeAdminNotification(
			context.Context,
			*system.AcknowledgeAdminNotificationReq,
			...grpc.CallOption,
		) (*system.AdminNotificationIncidentResp, error)
		ClaimDueAdminNotifications(
			context.Context,
			*system.ClaimDueAdminNotificationsReq,
			...grpc.CallOption,
		) (*system.ClaimDueAdminNotificationsResp, error)
		ReleaseAdminNotificationEscalation(
			context.Context,
			*system.ReleaseAdminNotificationEscalationReq,
			...grpc.CallOption,
		) (*system.RespBase, error)
	}

	NotificationPublisher interface {
		Publish(context.Context, string, any) error
	}

	NotificationStoreConfig struct {
		AckTimeout       time.Duration
		EscalationRepeat time.Duration
		EscalationMax    int64
		EscalationPoll   time.Duration
	}

	NotificationStore struct {
		client    NotificationSystemClient
		publisher NotificationPublisher
		config    NotificationStoreConfig
	}

	clientMessage struct {
		Action    string `json:"action"`
		EventType string `json:"eventType"`
		AlertKey  string `json:"alertKey"`
		TenantID  int64  `json:"tenantId"`
		Reason    string `json:"reason"`
	}

	clientMessageResponse struct {
		Action         string `json:"action"`
		OK             bool   `json:"ok"`
		EventType      string `json:"eventType,omitempty"`
		AlertKey       string `json:"alertKey,omitempty"`
		TenantID       int64  `json:"tenantId,omitempty"`
		AcknowledgedAt int64  `json:"acknowledgedAt,omitempty"`
		AcknowledgedBy int64  `json:"acknowledgedBy,omitempty"`
		Error          string `json:"error,omitempty"`
	}
)

func NewNotificationStore(
	client NotificationSystemClient,
	publisher NotificationPublisher,
	config NotificationStoreConfig,
) *NotificationStore {
	if config.AckTimeout < time.Minute {
		config.AckTimeout = 5 * time.Minute
	}
	if config.EscalationRepeat < time.Minute {
		config.EscalationRepeat = 5 * time.Minute
	}
	if config.EscalationMax <= 0 || config.EscalationMax > 10 {
		config.EscalationMax = 3
	}
	if config.EscalationPoll < time.Second {
		config.EscalationPoll = 15 * time.Second
	}
	return &NotificationStore{
		client:    client,
		publisher: publisher,
		config:    config,
	}
}

func (s *NotificationStore) RecordEvent(ctx context.Context, event notify.Event) error {
	if s == nil || s.client == nil || !isOperationalEvent(event.Type) {
		return nil
	}
	state := eventDataString(event.Data, "state")
	if state != "firing" && state != "resolved" {
		return errors.New("operational admin notification state is required")
	}
	alertKey := strings.TrimSpace(event.BizNo)
	if alertKey == "" {
		alertKey = eventDataString(event.Data, "alertKey")
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = s.client.RecordAdminNotification(ctx, &system.RecordAdminNotificationReq{
		TenantId:     event.TenantID,
		EventType:    event.Type,
		AlertKey:     alertKey,
		EventId:      event.ID,
		State:        state,
		Severity:     eventDataStringOr(event.Data, "severity", event.Level),
		Title:        event.Title,
		Message:      event.Message,
		Source:       event.Source,
		PayloadJson:  string(payload),
		EventTime:    event.CreatedAt,
		AckTimeoutMs: s.config.AckTimeout.Milliseconds(),
	})
	return err
}

func (s *NotificationStore) HandleClientMessage(
	ctx context.Context,
	connection *Connection,
	payload []byte,
) []byte {
	response := clientMessageResponse{Action: "ack_result"}
	var message clientMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		response.Error = "invalid notification acknowledgement"
		return marshalClientResponse(response)
	}
	message.Action = strings.TrimSpace(message.Action)
	message.EventType = strings.TrimSpace(message.EventType)
	message.AlertKey = strings.TrimSpace(message.AlertKey)
	message.Reason = strings.TrimSpace(message.Reason)
	response.EventType = message.EventType
	response.AlertKey = message.AlertKey
	response.TenantID = message.TenantID

	if s == nil || s.client == nil || connection == nil ||
		message.Action != "ack" || !isOperationalEvent(message.EventType) ||
		message.AlertKey == "" || message.Reason == "" {
		response.Error = "invalid notification acknowledgement"
		return marshalClientResponse(response)
	}
	if !connection.IsSystemAdmin && message.TenantID != connection.TenantId {
		response.Error = "notification tenant is forbidden"
		return marshalClientResponse(response)
	}
	if !connection.CanReceive(notify.Event{
		Type:     message.EventType,
		TenantID: message.TenantID,
	}) {
		response.Error = "notification permission is forbidden"
		return marshalClientResponse(response)
	}

	now := time.Now().UnixMilli()
	result, err := s.client.AcknowledgeAdminNotification(
		ctx,
		&system.AcknowledgeAdminNotificationReq{
			TenantId:       message.TenantID,
			EventType:      message.EventType,
			AlertKey:       message.AlertKey,
			UserId:         connection.UserId,
			Username:       connection.Username,
			Reason:         message.Reason,
			AcknowledgedAt: now,
		},
	)
	if err != nil || result == nil || result.Data == nil {
		response.Error = "notification acknowledgement failed"
		return marshalClientResponse(response)
	}
	response.OK = true
	response.AcknowledgedAt = result.Data.AcknowledgedAt
	response.AcknowledgedBy = result.Data.AcknowledgedBy
	return marshalClientResponse(response)
}

func (s *NotificationStore) RunEscalation(ctx context.Context) {
	if s == nil || s.client == nil || s.publisher == nil {
		return
	}
	ticker := time.NewTicker(s.config.EscalationPoll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			_ = s.EscalateOnce(ctx, now.UnixMilli())
		}
	}
}

func (s *NotificationStore) EscalateOnce(ctx context.Context, now int64) error {
	if s == nil || s.client == nil || s.publisher == nil {
		return errors.New("notification escalation dependencies are required")
	}
	result, err := s.client.ClaimDueAdminNotifications(
		ctx,
		&system.ClaimDueAdminNotificationsReq{
			Now:              now,
			RepeatIntervalMs: s.config.EscalationRepeat.Milliseconds(),
			MaxLevel:         s.config.EscalationMax,
			Limit:            100,
		},
	)
	if err != nil {
		return err
	}
	var publishErrors []error
	for _, incident := range result.GetData() {
		event, publishErr := escalatedEvent(incident, now)
		if publishErr == nil {
			publishErr = s.publisher.Publish(ctx, notify.Channel, event)
		}
		if publishErr == nil {
			continue
		}
		_, releaseErr := s.client.ReleaseAdminNotificationEscalation(
			ctx,
			&system.ReleaseAdminNotificationEscalationReq{
				Id:           incident.Id,
				ClaimedLevel: incident.EscalationLevel,
				RetryAt:      now + escalationRetryDelay.Milliseconds(),
				Now:          now,
			},
		)
		publishErrors = append(
			publishErrors,
			fmt.Errorf("escalate incident %d: %w", incident.Id, errors.Join(publishErr, releaseErr)),
		)
	}
	return errors.Join(publishErrors...)
}

func escalatedEvent(
	incident *system.AdminNotificationIncident,
	now int64,
) (notify.Event, error) {
	if incident == nil {
		return notify.Event{}, errors.New("notification incident is required")
	}
	var event notify.Event
	if err := json.Unmarshal([]byte(incident.PayloadJson), &event); err != nil {
		return notify.Event{}, err
	}
	if event.Data == nil {
		event.Data = make(map[string]any)
	}
	event.ID = fmt.Sprintf(
		"admin-escalation:%d:%d:%d",
		incident.Id,
		incident.EscalationLevel,
		incident.Version,
	)
	event.Level = notify.EventLevelError
	event.Title = fmt.Sprintf("[未确认升级 L%d] %s", incident.EscalationLevel, incident.Title)
	event.CreatedAt = now
	event.Data["state"] = "firing"
	event.Data["alertKey"] = incident.AlertKey
	event.Data["escalationLevel"] = incident.EscalationLevel
	event.Data["incidentId"] = incident.Id
	return event, nil
}

func isOperationalEvent(eventType string) bool {
	switch eventType {
	case "contract_reconciliation", "price_engine_input", "snapshot_outbox":
		return true
	default:
		return false
	}
}

func eventDataString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

func eventDataStringOr(data map[string]any, key string, fallback string) string {
	if value := eventDataString(data, key); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func marshalClientResponse(response clientMessageResponse) []byte {
	payload, _ := json.Marshal(response)
	return payload
}

var _ NotificationPublisher = (*mq.Publisher)(nil)
