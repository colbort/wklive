package ws

import (
	"sort"
	"strings"
	"time"

	"wklive/proto/chat"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func sortTransientSessions(list []*chat.ChatSession) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].GetLastMessageTime() == list[j].GetLastMessageTime() {
			return list[i].GetCreateTimes() > list[j].GetCreateTimes()
		}
		return list[i].GetLastMessageTime() > list[j].GetLastMessageTime()
	})
}

func filterTransientMessages(list []*chat.ChatMessage, senderType int64) []*chat.ChatMessage {
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
	return filtered
}

func limitTransientMessages(list []*chat.ChatMessage, limit int64) []*chat.ChatMessage {
	if limit <= 0 || int(limit) >= len(list) {
		return list
	}
	return list[len(list)-int(limit):]
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

func eventHasGuestSession(event *chat.ChatMessageEvent) bool {
	if event == nil {
		return false
	}
	if sessionIsGuest(event.GetSession()) {
		return true
	}
	return event.GetSessionEvent() != nil && sessionIsGuest(event.GetSessionEvent().GetSession())
}

func sessionIsGuest(session *chat.ChatSession) bool {
	if session == nil {
		return false
	}
	if session.GetIsGuest() {
		return true
	}
	return legacySessionIsGuest(session)
}

func legacySessionIsGuest(session *chat.ChatSession) bool {
	return structBool(session.GetExtJson(), "isGuest") || structBool(session.GetExtJson(), "is_guest")
}

func structBool(data *structpb.Struct, key string) bool {
	if data == nil {
		return false
	}
	value := data.GetFields()[key]
	if value == nil {
		return false
	}
	switch kind := value.GetKind().(type) {
	case *structpb.Value_BoolValue:
		return kind.BoolValue
	case *structpb.Value_StringValue:
		return strings.EqualFold(strings.TrimSpace(kind.StringValue), "true") || strings.TrimSpace(kind.StringValue) == "1"
	case *structpb.Value_NumberValue:
		return kind.NumberValue != 0
	default:
		return false
	}
}

func newTransientSession(sessionNo string) *chat.ChatSession {
	now := time.Now().UnixMilli()
	session := &chat.ChatSession{
		SessionNo:       sessionNo,
		Source:          chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
		Status:          chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT,
		Priority:        chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL,
		Title:           "访客",
		LastMessageTime: now,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	return session
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
	if src.GetExtJson() != nil {
		dst.ExtJson = src.GetExtJson()
	}
	if src.GetIsGuest() {
		dst.IsGuest = true
	}
	if src.GetAvatarUrl() != "" {
		dst.AvatarUrl = src.GetAvatarUrl()
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

func applyTransientEventSnapshot(session *chat.ChatSession, event *chat.ChatMessageEvent) {
	if session == nil || event == nil {
		return
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
}

func applyTransientEventMessage(session *chat.ChatSession, event *chat.ChatMessageEvent, fallback string) {
	if session == nil || event == nil {
		return
	}
	message := transientEventMessageText(event, fallback)
	if message != "" {
		session.LastMessage = message
	}
	if event.GetData() != nil && event.GetData().GetMessageNo() != "" {
		session.LastMessageNo = event.GetData().GetMessageNo()
	}
	if senderType := messageSenderType(event.GetData()); senderType != chat.ChatSenderType_CHAT_SENDER_TYPE_UNKNOWN {
		session.LastSenderType = senderType
	}
	timestamp := transientEventTimestamp(event)
	session.LastMessageTime = timestamp
	session.UpdateTimes = timestamp
}

func transientEventMessageText(event *chat.ChatMessageEvent, fallback string) string {
	message := strings.TrimSpace(fallback)
	if event.GetSessionEvent() != nil && strings.TrimSpace(event.GetSessionEvent().GetMessage()) != "" {
		message = strings.TrimSpace(event.GetSessionEvent().GetMessage())
	}
	if event.GetData() != nil && strings.TrimSpace(event.GetData().GetContent()) != "" {
		message = strings.TrimSpace(event.GetData().GetContent())
	}
	return message
}

func transientEventTimestamp(event *chat.ChatMessageEvent) int64 {
	return firstPositiveInt64(
		event.GetCreatedAt(),
		event.GetData().GetCreateTime(),
		event.GetSessionEvent().GetCreatedAt(),
		time.Now().UnixMilli(),
	)
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
		if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER && strings.TrimSpace(session.GetAvatarUrl()) == "" {
			session.AvatarUrl = strings.TrimSpace(msg.GetSender().GetAvatarUrl())
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

func applyTransientMessageStatus(session *chat.ChatSession, msg *chat.ChatMessage) {
	switch messageSenderType(msg) {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT
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

func isWaitingTransientSession(session *chat.ChatSession) bool {
	if session == nil {
		return false
	}
	status := session.GetStatus()
	return session.GetAgentId() == 0 &&
		(status == chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING ||
			status == chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
}

func firstPositiveInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
