package ws

import (
	"sort"
	"strings"
	"sync"
	"time"

	"wklive/proto/chat"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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
	if !s.isTransientEventLocked(sessionNo, event) {
		return
	}
	s.cleanupLocked(time.Now().UnixMilli())

	if event.GetType() == chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE {
		delete(s.sessions, sessionNo)
		delete(s.messages, sessionNo)
		return
	}

	session := s.sessions[sessionNo]
	if session == nil {
		session = newTransientSession(sessionNo)
		s.sessions[sessionNo] = session
	}

	if event.GetSession() != nil {
		mergeTransientSession(session, event.GetSession())
	}
	if event.GetSessionEvent() != nil {
		applyTransientSessionEvent(session, event.GetSessionEvent())
	}
	if event.GetQueue() != nil {
		applyTransientQueue(session, event.GetQueue())
	}
	if event.GetData() != nil && event.GetType() != chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE {
		applyTransientMessage(session, event.GetType(), event.GetData())
		if event.GetType() == chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE {
			s.appendMessageLocked(event.GetData())
		}
	}
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
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].GetLastMessageTime() == list[j].GetLastMessageTime() {
			return list[i].GetCreateTimes() > list[j].GetCreateTimes()
		}
		return list[i].GetLastMessageTime() > list[j].GetLastMessageTime()
	})
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
	filtered := make([]*chat.ChatMessage, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		if senderType > 0 && int64(messageSenderType(item)) != senderType {
			continue
		}
		filtered = append(filtered, cloneTransientMessage(item))
	}
	if limit <= 0 || int(limit) >= len(filtered) {
		return filtered
	}
	return filtered[len(filtered)-int(limit):]
}

func (s *transientSessionStore) appendMessageLocked(msg *chat.ChatMessage) {
	if msg == nil || msg.GetSessionNo() == "" || msg.GetMessageNo() == "" {
		return
	}
	list := s.messages[msg.GetSessionNo()]
	for _, item := range list {
		if item != nil && item.GetMessageNo() == msg.GetMessageNo() {
			return
		}
	}
	list = append(list, cloneTransientMessage(msg))
	if len(list) > transientMessageLimit {
		list = list[len(list)-transientMessageLimit:]
	}
	s.messages[msg.GetSessionNo()] = list
}

func cloneTransientSession(session *chat.ChatSession) *chat.ChatSession {
	if session == nil {
		return nil
	}
	return proto.Clone(session).(*chat.ChatSession)
}

func cloneTransientMessage(msg *chat.ChatMessage) *chat.ChatMessage {
	if msg == nil {
		return nil
	}
	return proto.Clone(msg).(*chat.ChatMessage)
}

func (s *transientSessionStore) cleanupLocked(now int64) {
	expiresBefore := now - transientSessionMaxAge.Milliseconds()
	for sessionNo, session := range s.sessions {
		if session.GetUpdateTimes() > 0 && session.GetUpdateTimes() < expiresBefore {
			delete(s.sessions, sessionNo)
			delete(s.messages, sessionNo)
		}
	}
}

func eventSessionNo(event *chat.ChatMessageEvent) string {
	if event.GetData() != nil {
		return event.GetData().GetSessionNo()
	}
	if event.GetSession() != nil {
		return event.GetSession().GetSessionNo()
	}
	if event.GetSessionEvent() != nil {
		return event.GetSessionEvent().GetSessionNo()
	}
	if event.GetQueue() != nil {
		return event.GetQueue().GetSessionNo()
	}
	return ""
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

func (s *transientSessionStore) isTransientEventLocked(sessionNo string, event *chat.ChatMessageEvent) bool {
	if event == nil || sessionNo == "" {
		return false
	}
	if _, ok := s.sessions[sessionNo]; ok {
		return true
	}
	if sessionIsGuest(event.GetSession()) {
		return true
	}
	if event.GetSessionEvent() != nil && sessionIsGuest(event.GetSessionEvent().GetSession()) {
		return true
	}
	return false
}

func sessionIsGuest(session *chat.ChatSession) bool {
	if session == nil || session.GetExtJson() == nil {
		return false
	}
	value := session.GetExtJson().GetFields()["isGuest"]
	return value != nil && value.GetBoolValue()
}

func newTransientSession(sessionNo string) *chat.ChatSession {
	now := time.Now().UnixMilli()
	return &chat.ChatSession{
		SessionNo:       sessionNo,
		Source:          chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
		Status:          chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT,
		Priority:        chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL,
		Title:           "访客",
		LastMessageTime: now,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
}

func mergeTransientSession(dst *chat.ChatSession, src *chat.ChatSession) {
	if dst == nil || src == nil {
		return
	}
	if src.GetSessionNo() != "" {
		dst.SessionNo = src.GetSessionNo()
	}
	if src.GetMerchantId() > 0 {
		dst.MerchantId = src.GetMerchantId()
	}
	if src.GetUserId() != 0 {
		dst.UserId = src.GetUserId()
	}
	if src.GetSource() != chat.ChatSessionSource_CHAT_SESSION_SOURCE_UNKNOWN {
		dst.Source = src.GetSource()
	}
	if src.GetStatus() != chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN {
		dst.Status = src.GetStatus()
	}
	if src.GetPriority() != chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_UNKNOWN {
		dst.Priority = src.GetPriority()
	}
	if src.GetAgentId() > 0 {
		dst.AgentId = src.GetAgentId()
	}
	if src.GetGroupId() > 0 {
		dst.GroupId = src.GetGroupId()
	}
	if strings.TrimSpace(src.GetTitle()) != "" {
		dst.Title = src.GetTitle()
	}
	if strings.TrimSpace(src.GetCategory()) != "" {
		dst.Category = src.GetCategory()
	}
	if strings.TrimSpace(src.GetLastMessage()) != "" {
		dst.LastMessage = src.GetLastMessage()
	}
	if src.GetLastSenderType() != chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN {
		dst.LastSenderType = src.GetLastSenderType()
	}
	if src.GetLastMessageTime() > 0 {
		dst.LastMessageTime = src.GetLastMessageTime()
	}
	if src.GetUserUnreadCount() > 0 {
		dst.UserUnreadCount = src.GetUserUnreadCount()
	}
	if src.GetAgentUnreadCount() > 0 {
		dst.AgentUnreadCount = src.GetAgentUnreadCount()
	}
	if src.GetCloseTime() > 0 {
		dst.CloseTime = src.GetCloseTime()
	}
	if strings.TrimSpace(src.GetCloseReason()) != "" {
		dst.CloseReason = src.GetCloseReason()
	}
	if strings.TrimSpace(src.GetLastMessageNo()) != "" {
		dst.LastMessageNo = src.GetLastMessageNo()
	}
	if src.GetCreateTimes() > 0 {
		dst.CreateTimes = src.GetCreateTimes()
	}
	if src.GetUpdateTimes() > 0 {
		dst.UpdateTimes = src.GetUpdateTimes()
	}
}

func applyTransientSessionEvent(session *chat.ChatSession, event *chat.ChatSessionEvent) {
	if session == nil || event == nil {
		return
	}
	if event.GetSession() != nil {
		mergeTransientSession(session, event.GetSession())
	}
	if event.GetMerchantId() > 0 {
		session.MerchantId = event.GetMerchantId()
	}
	if event.GetUserId() != 0 {
		session.UserId = event.GetUserId()
	}
	if event.GetAgentId() > 0 {
		session.AgentId = event.GetAgentId()
	}
	if event.GetStatus() != chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN {
		session.Status = event.GetStatus()
	}
	if event.GetCreatedAt() > 0 {
		session.UpdateTimes = event.GetCreatedAt()
	}
}

func applyTransientQueue(session *chat.ChatSession, queue *chat.ChatQueueInfo) {
	if session == nil || queue == nil {
		return
	}
	if queue.GetMerchantId() > 0 {
		session.MerchantId = queue.GetMerchantId()
	}
	if queue.GetUserId() != 0 {
		session.UserId = queue.GetUserId()
	}
	if queue.GetGroupId() > 0 {
		session.GroupId = queue.GetGroupId()
	}
	if queue.GetUpdateTimes() > 0 {
		session.UpdateTimes = queue.GetUpdateTimes()
	}
}

func applyTransientMessage(session *chat.ChatSession, eventType chat.ChatEventType, msg *chat.ChatMessage) {
	if session == nil || msg == nil {
		return
	}
	if msg.GetSender() != nil && msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER && msg.GetSender().GetId() != 0 {
		session.UserId = msg.GetSender().GetId()
	}
	if agentId := int64FromString(msg.GetAgentId()); agentId > 0 {
		session.AgentId = agentId
	}
	if msg.GetSender() != nil && msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT && msg.GetSender().GetId() > 0 {
		session.AgentId = msg.GetSender().GetId()
	}
	if msg.GetSender() != nil {
		if name := strings.TrimSpace(msg.GetSender().GetNickname()); session.GetTitle() == "" && name != "" {
			session.Title = name
		}
		if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
			ensureTransientUserExt(session, msg.GetSender().GetAvatarUrl())
		}
	}
	if content := strings.TrimSpace(msg.GetContent()); content != "" {
		session.LastMessage = content
	}
	if msg.GetMessageNo() != "" {
		session.LastMessageNo = msg.GetMessageNo()
	}
	if senderType := messageSenderType(msg); senderType != chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN {
		session.LastSenderType = senderType
	}
	if msg.GetCreateTime() > 0 {
		session.LastMessageTime = msg.GetCreateTime()
	}
	if msg.GetUpdateTime() > 0 {
		session.UpdateTimes = msg.GetUpdateTime()
	}

	switch eventType {
	case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_START:
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING
	case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE:
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED
		session.CloseTime = msg.GetCreateTime()
	default:
		applyTransientMessageStatus(session, msg)
	}
}

func ensureTransientUserExt(session *chat.ChatSession, avatarUrl string) {
	if session == nil || strings.TrimSpace(avatarUrl) == "" {
		return
	}
	if session.ExtJson != nil {
		if value := session.ExtJson.GetFields()["userAvatarUrl"].GetStringValue(); strings.TrimSpace(value) != "" {
			return
		}
	}
	ext, err := structpb.NewStruct(map[string]interface{}{
		"userAvatarUrl": strings.TrimSpace(avatarUrl),
	})
	if err != nil {
		return
	}
	session.ExtJson = ext
}

func applyTransientMessageStatus(session *chat.ChatSession, msg *chat.ChatMessage) {
	switch messageSenderType(msg) {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		if session.GetAgentId() > 0 {
			session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT
		} else {
			session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT
		}
		session.AgentUnreadCount++
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER
		session.UserUnreadCount++
	}
}

func messageSenderType(msg *chat.ChatMessage) chat.ChatSenderType {
	if msg == nil || msg.GetSender() == nil {
		return chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN
	}
	return msg.GetSender().GetType()
}

func matchTransientSession(session *chat.ChatSession, filter TransientSessionFilter) bool {
	if session == nil {
		return false
	}
	if filter.MerchantId > 0 && session.GetMerchantId() != filter.MerchantId {
		return false
	}
	if filter.UserId != 0 && session.GetUserId() != filter.UserId {
		return false
	}
	if filter.AgentId > 0 && session.GetAgentId() != filter.AgentId && session.GetAgentId() != 0 {
		return false
	}
	if filter.Status > 0 && int64(session.GetStatus()) != filter.Status {
		return false
	}
	return true
}
