package ws

import (
	"fmt"
	"sync"
	"time"

	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	transientSessionMaxAge = 24 * time.Hour
	transientMessageLimit  = 200
)

type TransientSessionFilter struct {
	MerchantId int64
	UserId     int64
	AgentId    int64
	Status     int64
}

type transientSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*chat.ChatSession
	messages map[string][]*chat.ChatMessage
}

func newTransientSessionStore() *transientSessionStore {
	return &transientSessionStore{
		sessions: make(map[string]*chat.ChatSession),
		messages: make(map[string][]*chat.ChatMessage),
	}
}

func (s *transientSessionStore) ApplyEvent(event *chat.ChatMessageEvent) {
	if s == nil || event == nil {
		return
	}
	sessionNo := eventSessionNo(event)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now().UnixMilli())

	if !s.shouldTrackEventLocked(sessionNo, event) {
		logx.Error("shouldTrackEventLocked ============ false")
		return
	}

	switch event.GetType() {
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN:
		s.applyUserJoinLocked(sessionNo, event)
	case chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE:
		s.applyUserLeaveLocked(sessionNo, event)
	case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
		s.applySessionCloseLocked(sessionNo, event)
	default:
		s.applySessionEventLocked(sessionNo, event)
	}
}

func (s *transientSessionStore) applySessionEventLocked(sessionNo string, event *chat.ChatMessageEvent) {
	session := s.getOrCreateSessionLocked(sessionNo)
	if session == nil {
		return
	}
	applyTransientEventSnapshot(session, event)
	if event.GetData() != nil && event.GetType() != chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE {
		applyTransientMessage(session, event.GetType(), event.GetData())
		if event.GetType() == chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE {
			s.appendMessageLocked(event.GetData())
		}
	}
}

func (s *transientSessionStore) applyUserJoinLocked(sessionNo string, event *chat.ChatMessageEvent) {
	session := s.getOrCreateSessionLocked(sessionNo)
	if session == nil {
		return
	}
	applyTransientEventSnapshot(session, event)
	if session.GetStatus() == chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN {
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING
	}
	if session.GetAgentId() == 0 {
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING
	}
	applyTransientEventMessage(session, event, "用户进入客服，等待接待")
}

func (s *transientSessionStore) applyUserLeaveLocked(sessionNo string, event *chat.ChatMessageEvent) {
	session := s.sessions[sessionNo]
	if session == nil {
		return
	}
	if isWaitingTransientSession(session) {
		s.deleteSessionLocked(sessionNo)
		return
	}
	applyTransientEventSnapshot(session, event)
	applyTransientEventMessage(session, event, "用户已离开")
}

func (s *transientSessionStore) applySessionCloseLocked(sessionNo string, event *chat.ChatMessageEvent) {
	session := s.getOrCreateSessionLocked(sessionNo)
	if session == nil {
		return
	}
	applyTransientEventSnapshot(session, event)
	if event.GetData() != nil {
		applyTransientMessage(session, event.GetType(), event.GetData())
		s.appendMessageLocked(event.GetData())
	}
	session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED
	if session.GetCloseTime() == 0 {
		session.CloseTime = firstPositiveInt64(event.GetCreatedAt(), session.GetUpdateTimes(), time.Now().UnixMilli())
	}
	applyTransientEventMessage(session, event, "本次会话已结束")
}

func (s *transientSessionStore) getOrCreateSessionLocked(sessionNo string) *chat.ChatSession {
	if sessionNo == "" {
		return nil
	}
	session := s.sessions[sessionNo]
	if session == nil {
		session = newTransientSession(sessionNo)
		s.sessions[sessionNo] = session
	}
	return session
}

func (s *transientSessionStore) deleteSessionLocked(sessionNo string) {
	delete(s.sessions, sessionNo)
	delete(s.messages, sessionNo)
}

func (s *transientSessionStore) List(filter TransientSessionFilter) []*chat.ChatSession {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*chat.ChatSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		if !matchTransientSession(session, filter) {
			continue
		}
		list = append(list, cloneTransientSession(session))
	}
	sortTransientSessions(list)
	return list
}

func (s *transientSessionStore) ListMessages(merchantId int64, sessionNo string, senderType int64, limit int64) []*chat.ChatMessage {
	if s == nil || sessionNo == "" {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[sessionNo]
	if session == nil {
		return nil
	}
	if merchantId > 0 && session.GetMerchantId() != merchantId {
		return nil
	}

	list := s.messages[sessionNo]
	if len(list) == 0 {
		return nil
	}
	return limitTransientMessages(filterTransientMessages(list, senderType), limit)
}

func (s *transientSessionStore) appendMessageLocked(msg *chat.ChatMessage) {
	if msg == nil || msg.GetSessionNo() == "" || msg.GetMessageNo() == "" {
		return
	}
	sessionNo := msg.GetSessionNo()
	list := s.messages[sessionNo]
	for _, item := range list {
		if item != nil && item.GetMessageNo() == msg.GetMessageNo() {
			return
		}
	}
	list = append(list, cloneTransientMessage(msg))
	if len(list) > transientMessageLimit {
		list = list[len(list)-transientMessageLimit:]
	}
	s.messages[sessionNo] = list
}

func (s *transientSessionStore) cleanupLocked(now int64) {
	expiresBefore := now - transientSessionMaxAge.Milliseconds()
	for sessionNo, session := range s.sessions {
		if session.GetUpdateTimes() > 0 && session.GetUpdateTimes() < expiresBefore {
			s.deleteSessionLocked(sessionNo)
		}
	}
}

func (s *transientSessionStore) IsTransientSession(sessionNo string) bool {
	if s == nil || sessionNo == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.sessions[sessionNo]
	return ok
}

func (s *transientSessionStore) shouldTrackEventLocked(sessionNo string, event *chat.ChatMessageEvent) bool {
	if event == nil || sessionNo == "" {
		fmt.Println("========================. 11")
		return false
	}
	if _, ok := s.sessions[sessionNo]; ok {
		return true
	}
	fmt.Println("========================. 22")
	return eventHasGuestSession(event)
}
