package ws

import (
	"encoding/json"
	"testing"

	"wklive/proto/chat"
)

func TestMarshalProtoJSONUsesEnumNumbers(t *testing.T) {
	payload, err := marshalProtoJSON((&chat.ChatWsResponse{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
	}).ProtoReflect())
	if err != nil {
		t.Fatalf("marshalProtoJSON() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got["eventType"] != float64(chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE) {
		t.Fatalf("eventType = %#v, want %d", got["eventType"], chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE)
	}
}
