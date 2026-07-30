package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	mq "wklive/common/mq/kafka"
	"wklive/common/notify"

	"github.com/gorilla/websocket"
)

const (
	notificationSubprotocol = "wklive-admin-notifications"
	defaultAdminAPIURL      = "http://127.0.0.1:8888"
	defaultKafkaBrokers     = "127.0.0.1:9092"
)

type (
	loginResponse struct {
		Code int32  `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	ackResponse struct {
		Action         string `json:"action"`
		OK             bool   `json:"ok"`
		EventType      string `json:"eventType"`
		AlertKey       string `json:"alertKey"`
		AcknowledgedAt int64  `json:"acknowledgedAt"`
		AcknowledgedBy int64  `json:"acknowledgedBy"`
		Error          string `json:"error"`
	}

	acceptanceResult struct {
		Result                   string `json:"result"`
		EventType                string `json:"eventType"`
		AlertKey                 string `json:"alertKey"`
		SourceEventAt            int64  `json:"sourceEventAt"`
		PublishedAt              int64  `json:"publishedAt"`
		InitialObservedAt        int64  `json:"initialObservedAt"`
		EscalationObservedAt     int64  `json:"escalationObservedAt"`
		EscalationLevel          int64  `json:"escalationLevel"`
		AcknowledgedAt           int64  `json:"acknowledgedAt"`
		AcknowledgedBy           int64  `json:"acknowledgedBy"`
		ResolvedObservedAt       int64  `json:"resolvedObservedAt"`
		MissingOriginRejected    bool   `json:"missingOriginRejected"`
		QueryTokenRejected       bool   `json:"queryTokenRejected"`
		FixedSubprotocolSelected bool   `json:"fixedSubprotocolSelected"`
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "admin notification acceptance failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()

	adminAPIURL := strings.TrimRight(envOr("ADMIN_API_URL", defaultAdminAPIURL), "/")
	websocketOrigin := strings.TrimRight(envOr("ADMIN_WS_ORIGIN", adminAPIURL), "/")
	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		return errors.New("ADMIN_USERNAME and ADMIN_PASSWORD are required")
	}

	token, err := login(ctx, adminAPIURL, username, password)
	if err != nil {
		return err
	}
	missingOriginRejected, queryTokenRejected, err := verifyWebSocketSecurity(
		ctx,
		adminAPIURL,
		websocketOrigin,
		token,
	)
	if err != nil {
		return err
	}
	connection, err := dialNotifications(ctx, adminAPIURL, websocketOrigin, token)
	if err != nil {
		return err
	}
	defer connection.Close()

	connected, err := readEvent(connection, 10*time.Second)
	if err != nil {
		return fmt.Errorf("read websocket greeting: %w", err)
	}
	if connected.ID != "connected" {
		return fmt.Errorf("unexpected websocket greeting %q", connected.ID)
	}

	publisher, err := mq.NewPublisher(mq.Config{
		Brokers:  splitNonEmpty(envOr("KAFKA_BROKERS", defaultKafkaBrokers)),
		ClientID: "admin-notification-acceptance",
	})
	if err != nil {
		return err
	}
	defer publisher.Close()

	now := time.Now()
	publishedAt := now.UnixMilli()
	sourceEventAt := now.Add(-10 * time.Minute).UnixMilli()
	eventType := envOr("ADMIN_NOTIFICATION_EVENT_TYPE", "snapshot_outbox")
	if !supportedEventType(eventType) {
		return fmt.Errorf("unsupported ADMIN_NOTIFICATION_EVENT_TYPE %q", eventType)
	}
	alertKey := fmt.Sprintf("acceptance-admin-ws-%d", now.UnixMilli())
	event := notify.Event{
		ID:        alertKey + "-firing",
		Type:      eventType,
		Level:     notify.EventLevelWarning,
		Title:     "Admin WS 验收告警",
		Message:   "验证持久化、升级、确认回执与恢复闭环",
		Source:    "admin-notification-acceptance",
		BizNo:     alertKey,
		CreatedAt: sourceEventAt,
		Data: map[string]any{
			"state":                 "firing",
			"severity":              "warning",
			"alertKey":              alertKey,
			"acceptancePublishedAt": publishedAt,
		},
	}
	if err = publisher.Publish(ctx, notify.Channel, event); err != nil {
		return fmt.Errorf("publish firing event: %w", err)
	}

	_, err = waitForEvent(connection, alertKey, "firing", 15*time.Second)
	if err != nil {
		return err
	}
	initialObservedAt := time.Now().UnixMilli()
	escalated, err := waitForEscalation(connection, alertKey, 35*time.Second)
	if err != nil {
		return err
	}
	escalationObservedAt := time.Now().UnixMilli()
	escalationLevel := eventDataInt64(escalated.Data, "escalationLevel")
	if escalationLevel < 1 {
		return errors.New("escalation level was not returned")
	}

	if err = connection.WriteJSON(map[string]any{
		"action":    "ack",
		"eventType": event.Type,
		"alertKey":  alertKey,
		"tenantId":  0,
		"reason":    "admin notification runtime acceptance",
	}); err != nil {
		return fmt.Errorf("send acknowledgement: %w", err)
	}
	acknowledgement, err := waitForAcknowledgement(connection, alertKey, 10*time.Second)
	if err != nil {
		return err
	}

	event.ID = alertKey + "-resolved"
	event.Level = notify.EventLevelInfo
	event.Title = "Admin WS 验收告警已恢复"
	event.Message = "运行态验收事件已确认恢复"
	event.CreatedAt = time.Now().UnixMilli()
	event.Data["state"] = "resolved"
	event.Data["severity"] = "info"
	if err = publisher.Publish(ctx, notify.Channel, event); err != nil {
		return fmt.Errorf("publish resolved event: %w", err)
	}
	if _, err = waitForEvent(connection, alertKey, "resolved", 15*time.Second); err != nil {
		return err
	}
	resolvedObservedAt := time.Now().UnixMilli()

	result, _ := json.Marshal(acceptanceResult{
		Result:                   "PASS",
		EventType:                event.Type,
		AlertKey:                 alertKey,
		SourceEventAt:            sourceEventAt,
		PublishedAt:              publishedAt,
		InitialObservedAt:        initialObservedAt,
		EscalationObservedAt:     escalationObservedAt,
		EscalationLevel:          escalationLevel,
		AcknowledgedAt:           acknowledgement.AcknowledgedAt,
		AcknowledgedBy:           acknowledgement.AcknowledgedBy,
		ResolvedObservedAt:       resolvedObservedAt,
		MissingOriginRejected:    missingOriginRejected,
		QueryTokenRejected:       queryTokenRejected,
		FixedSubprotocolSelected: connection.Subprotocol() == notificationSubprotocol,
	})
	fmt.Println(string(result))
	return nil
}

func login(ctx context.Context, baseURL, username, password string) (string, error) {
	payload, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/admin/system/auth/login",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("admin login: %w", err)
	}
	defer response.Body.Close()

	var result loginResponse
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode admin login: %w", err)
	}
	if response.StatusCode != http.StatusOK ||
		(result.Code != 0 && result.Code != http.StatusOK) ||
		result.Data.Token == "" {
		return "", fmt.Errorf(
			"admin login rejected: http=%d code=%d msg=%s",
			response.StatusCode,
			result.Code,
			result.Msg,
		)
	}
	return result.Data.Token, nil
}

func dialNotifications(
	ctx context.Context,
	adminAPIURL string,
	origin string,
	token string,
) (*websocket.Conn, error) {
	base, err := url.Parse(adminAPIURL)
	if err != nil {
		return nil, err
	}
	scheme := "ws"
	if base.Scheme == "https" {
		scheme = "wss"
	}
	wsURL := url.URL{
		Scheme: scheme,
		Host:   base.Host,
		Path:   "/admin/ws/notifications",
	}
	dialer := websocket.Dialer{
		Subprotocols: []string{
			notificationSubprotocol,
			"bearer." + token,
		},
		HandshakeTimeout: 10 * time.Second,
	}
	headers := http.Header{}
	headers.Set("Origin", origin)
	connection, response, err := dialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("dial admin websocket: http=%d: %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("dial admin websocket: %w", err)
	}
	if connection.Subprotocol() != notificationSubprotocol {
		connection.Close()
		return nil, fmt.Errorf("unexpected websocket subprotocol %q", connection.Subprotocol())
	}
	return connection, nil
}

func verifyWebSocketSecurity(
	ctx context.Context,
	adminAPIURL string,
	origin string,
	token string,
) (bool, bool, error) {
	base, err := url.Parse(adminAPIURL)
	if err != nil {
		return false, false, err
	}
	scheme := "ws"
	if base.Scheme == "https" {
		scheme = "wss"
	}
	wsURL := url.URL{
		Scheme: scheme,
		Host:   base.Host,
		Path:   "/admin/ws/notifications",
	}

	missingOriginDialer := websocket.Dialer{
		Subprotocols: []string{
			notificationSubprotocol,
			"bearer." + token,
		},
		HandshakeTimeout: 10 * time.Second,
	}
	connection, response, dialErr := missingOriginDialer.DialContext(
		ctx,
		wsURL.String(),
		nil,
	)
	if connection != nil {
		connection.Close()
	}
	missingOriginRejected := dialErr != nil &&
		response != nil &&
		response.StatusCode == http.StatusForbidden
	if !missingOriginRejected {
		return false, false, errors.New("websocket without Origin was not rejected")
	}

	queryTokenDialer := websocket.Dialer{
		Subprotocols:     []string{notificationSubprotocol},
		HandshakeTimeout: 10 * time.Second,
	}
	headers := http.Header{}
	headers.Set("Origin", origin)
	queryURL := wsURL
	query := queryURL.Query()
	query.Set("token", token)
	queryURL.RawQuery = query.Encode()
	connection, response, dialErr = queryTokenDialer.DialContext(
		ctx,
		queryURL.String(),
		headers,
	)
	if connection != nil {
		connection.Close()
	}
	queryTokenRejected := dialErr != nil &&
		response != nil &&
		response.StatusCode == http.StatusUnauthorized
	if !queryTokenRejected {
		return true, false, errors.New("websocket URL token was not rejected")
	}
	return true, true, nil
}

func waitForEvent(
	connection *websocket.Conn,
	alertKey string,
	state string,
	timeout time.Duration,
) (notify.Event, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		event, err := readEvent(connection, time.Until(deadline))
		if err != nil {
			return notify.Event{}, err
		}
		if event.BizNo == alertKey && eventDataString(event.Data, "state") == state {
			return event, nil
		}
	}
	return notify.Event{}, fmt.Errorf("timed out waiting for %s event", state)
}

func waitForEscalation(
	connection *websocket.Conn,
	alertKey string,
	timeout time.Duration,
) (notify.Event, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		event, err := readEvent(connection, time.Until(deadline))
		if err != nil {
			return notify.Event{}, err
		}
		if event.BizNo == alertKey && eventDataInt64(event.Data, "escalationLevel") > 0 {
			return event, nil
		}
	}
	return notify.Event{}, errors.New("timed out waiting for escalation event")
}

func waitForAcknowledgement(
	connection *websocket.Conn,
	alertKey string,
	timeout time.Duration,
) (ackResponse, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		connection.SetReadDeadline(deadline)
		_, payload, err := connection.ReadMessage()
		if err != nil {
			return ackResponse{}, err
		}
		var acknowledgement ackResponse
		if json.Unmarshal(payload, &acknowledgement) != nil ||
			acknowledgement.Action != "ack_result" ||
			acknowledgement.AlertKey != alertKey {
			continue
		}
		if !acknowledgement.OK {
			return ackResponse{}, fmt.Errorf(
				"notification acknowledgement rejected: %s",
				acknowledgement.Error,
			)
		}
		return acknowledgement, nil
	}
	return ackResponse{}, errors.New("timed out waiting for acknowledgement receipt")
}

func readEvent(connection *websocket.Conn, timeout time.Duration) (notify.Event, error) {
	connection.SetReadDeadline(time.Now().Add(timeout))
	_, payload, err := connection.ReadMessage()
	if err != nil {
		return notify.Event{}, err
	}
	var event notify.Event
	if err = json.Unmarshal(payload, &event); err != nil {
		return notify.Event{}, err
	}
	return event, nil
}

func eventDataString(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

func eventDataInt64(data map[string]any, key string) int64 {
	switch value := data[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	default:
		return 0
	}
}

func splitNonEmpty(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func supportedEventType(eventType string) bool {
	switch eventType {
	case "snapshot_outbox", "price_engine_input", "contract_reconciliation":
		return true
	default:
		return false
	}
}

func envOr(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
