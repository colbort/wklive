package ws

import (
	"testing"
	"time"

	"wklive/proto/chat"
)

func TestTransientStoreApplyGuestUserJoin(t *testing.T) {
	store := newTransientSessionStore()
	now := time.Now().UnixMilli()

	store.ApplyEvent(&chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		CreatedAt: now,
		Session: &chat.ChatSession{
			SessionNo:  "CS-GUEST",
			MerchantId: 2,
			UserId:     1,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			Title:      "访客",
			IsGuest:    true,
			AvatarUrl:  "https://example.com/avatar.png",
		},
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  "CS-GUEST",
			MerchantId: 2,
			UserId:     1,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			Message:    "用户进入客服，等待接待",
			CreatedAt:  now,
		},
	})

	sessions := store.List(TransientSessionFilter{
		MerchantId: 2,
		Status:     int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING),
	})
	if len(sessions) != 1 {
		t.Fatalf("guest user join should create transient session, got %d", len(sessions))
	}
	if sessions[0].GetSessionNo() != "CS-GUEST" {
		t.Fatalf("sessionNo = %q", sessions[0].GetSessionNo())
	}
}

func TestTransientStoreIgnoreNonGuestUserJoin(t *testing.T) {
	store := newTransientSessionStore()

	store.ApplyEvent(&chat.ChatMessageEvent{
		Type: chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		Session: &chat.ChatSession{
			SessionNo:  "CS-LOGIN",
			MerchantId: 2,
			UserId:     1,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			IsGuest:    false,
		},
	})

	if store.IsTransientSession("CS-LOGIN") {
		t.Fatal("non-guest user join should not create transient session")
	}
}

func TestTransientStoreKeepsWaitingGuestOnUserLeave(t *testing.T) {
	store := newTransientSessionStore()
	now := time.Now().UnixMilli()

	store.ApplyEvent(&chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		CreatedAt: now,
		Session: &chat.ChatSession{
			SessionNo:  "CS-GUEST",
			MerchantId: 2,
			UserId:     1,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			Title:      "访客",
			IsGuest:    true,
		},
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  "CS-GUEST",
			MerchantId: 2,
			UserId:     1,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			Message:    "用户进入客服，等待接待",
			CreatedAt:  now,
		},
	})
	store.ApplyEvent(&chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
		CreatedAt: now + 1,
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  "CS-GUEST",
			MerchantId: 2,
			UserId:     1,
			Message:    "用户已离开客服页面",
			CreatedAt:  now + 1,
		},
	})

	sessions := store.List(TransientSessionFilter{
		MerchantId: 2,
		Status:     int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING),
	})
	if len(sessions) != 1 {
		t.Fatalf("guest user leave should keep waiting transient session, got %d", len(sessions))
	}
	if sessions[0].GetLastMessage() != "用户已离开客服页面" {
		t.Fatalf("lastMessage = %q", sessions[0].GetLastMessage())
	}
}
