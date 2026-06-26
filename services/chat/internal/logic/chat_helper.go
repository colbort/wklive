package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultAgentMaxSessionCount = 10
	defaultDeviceID             = ""
	sessionNoInsertAttempts     = 3
)

var sequence uint64

func okBase() *common.RespBase {
	return helper.OkResp()
}

func badBase(msg string) *common.RespBase {
	return helper.GetErrResp(400, msg)
}

func notFoundBase(msg string) *common.RespBase {
	return helper.GetErrResp(404, msg)
}

func errorBase(err error) *common.RespBase {
	if err == nil {
		return helper.FailResp()
	}
	return helper.GetErrResp(500, err.Error())
}

func nowMillis() int64 {
	return time.Now().UnixMilli()
}

func merchantIDFromMetadata(ctx context.Context) (int64, *common.RespBase, error) {
	merchantID, err := utils.GetMerchantIdFromMd(ctx)
	if err != nil || merchantID <= 0 {
		return 0, badBase("merchant_id is required"), nil
	}
	return merchantID, nil, nil
}

func chatAppIdentityFromMetadata(ctx context.Context) (int64, int64, *common.RespBase, error) {
	merchantID, base, err := merchantIDFromMetadata(ctx)
	if base != nil || err != nil {
		return 0, 0, base, err
	}
	userID, err := utils.GetUserIdFromMd(ctx)
	if err != nil || userID == 0 {
		return 0, 0, badBase("user_id is required"), nil
	}
	return merchantID, userID, nil, nil
}

func nextNo(prefix string) string {
	n := atomic.AddUint64(&sequence, 1)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%06d%s", prefix, time.Now().UnixMilli(), n%1000000, hex.EncodeToString(b))
}

func pageInput(page *common.PageReq) (int64, int64) {
	return pageutil.Input(page)
}

func offsetBase(cursor, limit int64, size int, total int64) *common.RespBase {
	if cursor < 0 {
		cursor = 0
	}
	if limit <= 0 {
		limit = pageutil.NormalizeLimit(limit)
	}
	nextCursor := cursor + int64(size)
	hasNext := nextCursor < total
	prevCursor := cursor - limit
	if prevCursor < 0 {
		prevCursor = 0
	}
	return helper.OkWithOthers(total, hasNext, cursor > 0, nextCursor, prevCursor)
}

func messageNextCursor(list []*models.ChatMessage) int64 {
	if len(list) == 0 {
		return 0
	}
	return list[len(list)-1].CreateTimes
}

func normalizeMessageType(value chat.ChatMessageType) chat.ChatMessageType {
	if value == chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNKNOWN {
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT
	}
	return value
}

func normalizeSource(value chat.ChatSessionSource) chat.ChatSessionSource {
	if value == chat.ChatSessionSource_CHAT_SESSION_SOURCE_UNKNOWN {
		return chat.ChatSessionSource_CHAT_SESSION_SOURCE_APP
	}
	return value
}

func normalizePriority(value chat.ChatSessionPriority) chat.ChatSessionPriority {
	if value == chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_UNKNOWN {
		return chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL
	}
	return value
}

func normalizeAssignType(value chat.ChatAssignType) chat.ChatAssignType {
	if value == chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN {
		return chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL
	}
	return value
}

func isWorkOrderFinished(status int64) bool {
	return status == 3 || status == 4
}

func validateSessionKey(merchantID int64, sessionNo string) error {
	if merchantID <= 0 {
		return fmt.Errorf("merchant_id is required")
	}
	if strings.TrimSpace(sessionNo) == "" {
		return fmt.Errorf("session_no is required")
	}
	return nil
}

func trimSummary(content, mediaName, mediaURL string) string {
	summary := strings.TrimSpace(content)
	if summary == "" {
		summary = strings.TrimSpace(mediaName)
	}
	if summary == "" {
		summary = strings.TrimSpace(mediaURL)
	}
	if len([]rune(summary)) <= 200 {
		return summary
	}
	return string([]rune(summary)[:200])
}

func structToNullString(st *structpb.Struct) sql.NullString {
	if st == nil {
		return sql.NullString{}
	}
	bs, err := json.Marshal(st.AsMap())
	if err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(bs), Valid: true}
}

func nullStringToStruct(ns sql.NullString) *structpb.Struct {
	if !ns.Valid || strings.TrimSpace(ns.String) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil
	}
	st, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return st
}

func mapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}
	st, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return st
}

func toProtoAgent(data *models.TChatAgent) *chat.ChatAgent {
	if data == nil {
		return nil
	}
	return &chat.ChatAgent{
		Id:                  data.Id,
		MerchantId:          data.MerchantId,
		ChatUserId:          data.ChatUserId,
		AgentNo:             data.AgentNo,
		WelcomeMessage:      data.WelcomeMessage,
		Status:              chat.ChatAgentStatus(data.Status),
		AutoOnline:          common.YesNo(data.AutoOnline),
		MaxSessionCount:     int32(data.MaxSessionCount),
		CurrentSessionCount: int32(data.CurrentSessionCount),
		LastActiveTime:      data.LastActiveTime,
		Remark:              data.Remark,
		CreateTimes:         data.CreateTimes,
		UpdateTimes:         data.UpdateTimes,
		GroupId:             data.GroupId,
	}
}

func toProtoAgents(list []*models.TChatAgent) []*chat.ChatAgent {
	resp := make([]*chat.ChatAgent, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoAgent(item))
	}
	return resp
}

func toProtoChatGroups(list []*models.TChatGroup) []*chat.ChatGroup {
	resp := make([]*chat.ChatGroup, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoChatGroup(item))
	}
	return resp
}

func toProtoChatCategory(data *models.TChatCategory) *chat.ChatCategory {
	if data == nil {
		return nil
	}
	return &chat.ChatCategory{
		Id:           data.Id,
		MerchantId:   data.MerchantId,
		ParentId:     data.ParentId,
		CategoryCode: data.CategoryCode,
		CategoryName: data.CategoryName,
		GroupId:      data.GroupId,
		Enabled:      common.Enable(data.Enabled),
		Sort:         int32(data.Sort),
		Remark:       data.Remark,
		CreateTimes:  data.CreateTimes,
		UpdateTimes:  data.UpdateTimes,
	}
}

func toProtoChatCategories(list []*models.TChatCategory) []*chat.ChatCategory {
	resp := make([]*chat.ChatCategory, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoChatCategory(item))
	}
	return resp
}

func nullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}

func stringFromNull(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func toProtoChatQuickReply(data *models.TChatQuickReply) *chat.ChatQuickReply {
	if data == nil {
		return nil
	}
	return &chat.ChatQuickReply{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		AgentId:     data.AgentId,
		CategoryId:  data.CategoryId,
		Title:       data.Title,
		Content:     stringFromNull(data.Content),
		Enabled:     common.Enable(data.Enabled),
		Sort:        int32(data.Sort),
		Remark:      data.Remark,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func toProtoChatQuickReplies(list []*models.TChatQuickReply) []*chat.ChatQuickReply {
	resp := make([]*chat.ChatQuickReply, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoChatQuickReply(item))
	}
	return resp
}

func toProtoChatWorkOrder(data *models.TChatWorkOrder) *chat.ChatWorkOrder {
	if data == nil {
		return nil
	}
	return &chat.ChatWorkOrder{
		Id:            data.Id,
		MerchantId:    data.MerchantId,
		WorkOrderNo:   data.WorkOrderNo,
		SessionNo:     data.SessionNo,
		UserId:        data.UserId,
		AgentId:       data.AgentId,
		GroupId:       data.GroupId,
		Title:         data.Title,
		Content:       stringFromNull(data.Content),
		ContactName:   data.ContactName,
		ContactMobile: data.ContactMobile,
		ContactEmail:  data.ContactEmail,
		Priority:      chat.ChatSessionPriority(data.Priority),
		Status:        int32(data.Status),
		HandlerId:     data.HandlerId,
		HandleResult:  data.HandleResult,
		FinishTime:    data.FinishTime,
		Remark:        data.Remark,
		CreateTimes:   data.CreateTimes,
		UpdateTimes:   data.UpdateTimes,
	}
}

func toProtoChatWorkOrders(list []*models.TChatWorkOrder) []*chat.ChatWorkOrder {
	resp := make([]*chat.ChatWorkOrder, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoChatWorkOrder(item))
	}
	return resp
}

func toProtoSatisfaction(data *models.TChatSatisfaction) *chat.ChatSatisfaction {
	if data == nil {
		return nil
	}
	return &chat.ChatSatisfaction{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		SessionNo:   data.SessionNo,
		UserId:      data.UserId,
		AgentId:     data.AgentId,
		Score:       int32(data.Score),
		Content:     data.Content,
		Tags:        data.Tags,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func toProtoMerchant(data *models.TChatMerchantInfo) *chat.ChatMerchant {
	if data == nil {
		return nil
	}
	return &chat.ChatMerchant{
		MerchantId:  data.MerchantId,
		ApiKey:      data.ApiKey,
		Enabled:     common.Enable(data.Enabled),
		ExpireTime:  data.ExpireTime,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func toProtoSession(data *models.TChatSession) *chat.ChatSession {
	if data == nil {
		return nil
	}
	return &chat.ChatSession{
		Id:               data.Id,
		SessionNo:        data.SessionNo,
		MerchantId:       data.MerchantId,
		UserId:           data.UserId,
		Source:           chat.ChatSessionSource(data.Source),
		Status:           chat.ChatSessionStatus(data.Status),
		Priority:         chat.ChatSessionPriority(data.Priority),
		AgentId:          data.AgentId,
		Title:            data.Title,
		Category:         data.Category,
		LastMessage:      data.LastMessage,
		LastSenderType:   chat.ChatSenderType(data.LastSenderType),
		LastMessageTime:  data.LastMessageTime,
		UserUnreadCount:  int32(data.UserUnreadCount),
		AgentUnreadCount: int32(data.AgentUnreadCount),
		CloseTime:        data.CloseTime,
		CloseReason:      data.CloseReason,
		ExtJson:          nullStringToStruct(data.ExtJson),
		CreateTimes:      data.CreateTimes,
		UpdateTimes:      data.UpdateTimes,
		GroupId:          data.GroupId,
		LastMessageNo:    data.LastMessageNo,
	}
}

func toProtoSessions(list []*models.TChatSession) []*chat.ChatSession {
	resp := make([]*chat.ChatSession, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoSession(item))
	}
	return resp
}

func toProtoQueueInfo(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*chat.ChatQueueInfo, error) {
	if session == nil {
		return nil, nil
	}
	position, waitingCount, err := svcCtx.ChatSessionModel.CountWaitingPosition(ctx, session)
	if err != nil {
		return nil, err
	}
	message := "正在排队，客服会尽快接入。"
	if position > 0 {
		if position == 1 {
			message = "您是当前队列第 1 位，客服即将接入。"
		} else {
			message = fmt.Sprintf("正在排队，您前面还有 %d 人。", position-1)
		}
	}
	if session.AgentId > 0 ||
		session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) ||
		session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) {
		message = "客服已接入。"
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		message = "本次会话已结束。"
	}
	return &chat.ChatQueueInfo{
		MerchantId:          session.MerchantId,
		SessionNo:           session.SessionNo,
		UserId:              session.UserId,
		GroupId:             session.GroupId,
		Position:            int32(position),
		WaitingCount:        int32(waitingCount),
		EstimateWaitSeconds: estimateWaitSeconds(position),
		Message:             message,
		UpdateTimes:         nowMillis(),
	}, nil
}

func estimateWaitSeconds(position int64) int64 {
	if position <= 1 {
		return 0
	}
	return (position - 1) * 60
}

func toProtoMessage(data *models.ChatMessage) *chat.ChatMessage {
	if data == nil {
		return nil
	}
	return &chat.ChatMessage{
		MessageNo:   data.MessageNo,
		SessionNo:   data.SessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		MessageType: chat.ChatMessageType(data.MessageType),
		Sender:      toProtoMessageSender(data),
		Content:     data.Content,
		Url:         data.MediaUrl,
		FileName:    data.MediaName,
		FileSize:    data.MediaSize,
		MimeType:    data.MediaMime,
		Status:      chat.ChatMessageStatus(data.Status),
		AgentId:     strconv.FormatInt(data.AgentId, 10),
		Extra:       messagePayloadJSON(data.Payload),
		CreateTime:  data.CreateTimes,
		UpdateTime:  data.UpdateTimes,
	}
}

func newEventSystemMessage(session *models.TChatSession, content string) *chat.ChatMessage {
	if session == nil {
		return nil
	}
	now := nowMillis()
	return &chat.ChatMessage{
		MessageNo:   nextNo("GM"),
		SessionNo:   session.SessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM,
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Content:     strings.TrimSpace(content),
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		AgentId:     strconv.FormatInt(session.AgentId, 10),
		CreateTime:  now,
		UpdateTime:  now,
		Sender: &chat.ChatMessageUser{
			Id:       session.AgentId,
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: "系统",
		},
	}
}

func toProtoMessages(list []*models.ChatMessage) []*chat.ChatMessage {
	resp := make([]*chat.ChatMessage, 0, len(list))
	for _, item := range list {
		resp = append(resp, toProtoMessage(item))
	}
	return resp
}

func toProtoMessageSender(data *models.ChatMessage) *chat.ChatMessageUser {
	if data == nil {
		return nil
	}
	if data.Sender != nil {
		return &chat.ChatMessageUser{
			Id:        data.Sender.Id,
			Type:      chat.ChatSenderType(data.Sender.Type),
			Nickname:  data.Sender.Nickname,
			AvatarUrl: data.Sender.AvatarUrl,
		}
	}
	return &chat.ChatMessageUser{
		Type: chat.ChatSenderType(data.SenderType),
	}
}

func messagePayloadJSON(payload map[string]interface{}) string {
	if len(payload) == 0 {
		return ""
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func toProtoUser(data *models.TChatUser) *chat.ChatUser {
	if data == nil {
		return nil
	}
	return &chat.ChatUser{
		Id:            data.Id,
		MerchantId:    data.MerchantId,
		UserType:      chat.ChatUserType(data.UserType),
		IsOwner:       common.YesNo(data.IsOwner),
		Username:      data.Username,
		Nickname:      data.Nickname,
		AvatarUrl:     data.AvatarUrl,
		Mobile:        data.Mobile,
		Email:         data.Email,
		Enabled:       common.Enable(data.Enabled),
		LastLoginTime: data.LastLoginTime,
		Remark:        data.Remark,
		CreateTimes:   data.CreateTimes,
		UpdateTimes:   data.UpdateTimes,
		Password:      data.Password,
		LastLoginIp:   data.LastLoginIp,
	}
}

func getSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantID int64, sessionNo string) (*models.TChatSession, *common.RespBase, error) {
	if err := validateSessionKey(merchantID, sessionNo); err != nil {
		return nil, badBase(err.Error()), nil
	}
	data, err := svcCtx.ChatSessionModel.FindOneBySessionNo(ctx, sessionNo)
	if err == models.ErrNotFound {
		return nil, notFoundBase("chat session not found"), nil
	}
	if err != nil {
		return nil, nil, err
	}
	if data.MerchantId != merchantID {
		return nil, notFoundBase("chat session not found"), nil
	}
	return data, nil, nil
}

func ensureOpenSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantID, userID int64, source chat.ChatSessionSource, title, category string, priority chat.ChatSessionPriority, ext *structpb.Struct) (*models.TChatSession, bool, error) {
	sessionSource := normalizeSource(source)
	data, err := svcCtx.ChatSessionModel.FindLatestByUserSource(ctx, merchantID, userID, int64(sessionSource))
	if err == nil {
		now := nowMillis()
		changed := false
		shouldNotifyQueue := false
		if data.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			data.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
			data.AgentId = 0
			data.CloseTime = 0
			data.CloseReason = ""
			changed = true
			shouldNotifyQueue = true
		}
		if data.Source == 0 {
			data.Source = int64(sessionSource)
			changed = true
		}
		if strings.TrimSpace(data.Title) == "" && strings.TrimSpace(title) != "" {
			data.Title = strings.TrimSpace(title)
			changed = true
		}
		if strings.TrimSpace(data.Category) == "" && strings.TrimSpace(category) != "" {
			data.Category = strings.TrimSpace(category)
			changed = true
		}
		if ext != nil {
			data.ExtJson = structToNullString(ext)
			changed = true
		}
		if changed {
			data.UpdateTimes = now
			if err := svcCtx.ChatSessionModel.Update(ctx, data); err != nil {
				return nil, false, err
			}
		}
		return data, shouldNotifyQueue, nil
	}
	if err != models.ErrNotFound {
		return nil, false, err
	}

	now := nowMillis()
	for attempt := 0; attempt < sessionNoInsertAttempts; attempt++ {
		data = &models.TChatSession{
			SessionNo:       nextNo("CS"),
			MerchantId:      merchantID,
			UserId:          userID,
			Source:          int64(sessionSource),
			Status:          int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING),
			Priority:        int64(normalizePriority(priority)),
			Title:           strings.TrimSpace(title),
			Category:        strings.TrimSpace(category),
			LastMessageTime: now,
			ExtJson:         structToNullString(ext),
			CreateTimes:     now,
			UpdateTimes:     now,
		}
		result, err := svcCtx.ChatSessionModel.Insert(ctx, data)
		if err == nil {
			if id, err := result.LastInsertId(); err == nil {
				data.Id = id
			}
			return data, true, nil
		}
		if !isDuplicateKey(err) {
			return nil, false, err
		}
	}
	return nil, false, fmt.Errorf("failed to generate unique session_no")
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}

func newMessage(session *models.TChatSession, senderType chat.ChatSenderType, senderID int64, senderName, senderAvatarURL string, messageType chat.ChatMessageType, content, mediaURL, mediaName, mediaMIME string, mediaSize int64, payload *structpb.Struct) *models.ChatMessage {
	now := nowMillis()
	senderName = normalizeSenderName(senderType, senderID, senderName)
	msg := &models.ChatMessage{
		MessageNo:  nextNo("CM"),
		SessionNo:  session.SessionNo,
		MerchantId: session.MerchantId,
		UserId:     session.UserId,
		AgentId:    session.AgentId,
		SenderType: int64(senderType),
		Sender: &models.ChatMessageSender{
			Id:        senderID,
			Type:      int64(senderType),
			Nickname:  senderName,
			AvatarUrl: strings.TrimSpace(senderAvatarURL),
		},
		MessageType: int64(normalizeMessageType(messageType)),
		Content:     strings.TrimSpace(content),
		MediaUrl:    strings.TrimSpace(mediaURL),
		MediaName:   strings.TrimSpace(mediaName),
		MediaMime:   strings.TrimSpace(mediaMIME),
		MediaSize:   mediaSize,
		Status:      int64(chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT),
		CreateTimes: now,
		UpdateTimes: now,
	}
	if payload != nil {
		msg.Payload = payload.AsMap()
	}
	return msg
}

func normalizeSenderName(senderType chat.ChatSenderType, senderID int64, senderName string) string {
	senderName = strings.TrimSpace(senderName)
	if senderName != "" {
		return senderName
	}
	switch senderType {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		return fmt.Sprintf("用户%d", senderID)
	case chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM:
		return "system"
	default:
		return ""
	}
}

func fillMessageSender(ctx context.Context, svcCtx *svc.ServiceContext, msg *models.ChatMessage) {
	if msg == nil {
		return
	}
	if msg.Sender == nil {
		msg.Sender = &models.ChatMessageSender{
			Type: msg.SenderType,
		}
	}
	switch chat.ChatSenderType(msg.SenderType) {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		fillAgentSenderSnapshot(ctx, svcCtx, msg)
	case chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM:
		if msg.Sender.Nickname == "" {
			msg.Sender.Nickname = "system"
		}
	}
}

func fillAgentSenderSnapshot(ctx context.Context, svcCtx *svc.ServiceContext, msg *models.ChatMessage) {
	if msg == nil || msg.Sender == nil || msg.Sender.Id <= 0 {
		return
	}
	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, msg.Sender.Id)
	if err != nil || agent == nil || agent.ChatUserId <= 0 {
		return
	}
	user, err := svcCtx.ChatUserModel.FindOne(ctx, agent.ChatUserId)
	if err != nil || user == nil {
		return
	}
	msg.Sender.Id = agent.Id
	msg.Sender.Type = msg.SenderType
	msg.Sender.Nickname = user.Nickname
	msg.Sender.AvatarUrl = user.AvatarUrl
}

func sendMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage) (*models.ChatMessage, error) {
	model := svcCtx.ChatMessageFactory.New(session.MerchantId)
	if model == nil {
		return nil, fmt.Errorf("invalid merchant_id: %d", session.MerchantId)
	}
	fillMessageSender(ctx, svcCtx, msg)
	if err := model.Insert(ctx, msg); err != nil {
		return nil, err
	}

	now := msg.CreateTimes
	session.LastMessageNo = msg.MessageNo
	session.LastMessage = trimSummary(msg.Content, msg.MediaName, msg.MediaUrl)
	session.LastSenderType = msg.SenderType
	session.LastMessageTime = now
	session.UpdateTimes = now
	switch chat.ChatSenderType(msg.SenderType) {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		session.AgentUnreadCount++
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
		}
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		session.UserUnreadCount++
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER)
		}
	case chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM:
		session.UserUnreadCount++
	}
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, err
	}
	publishMessageEvent(ctx, svcCtx, session, msg)
	if chat.ChatSenderType(msg.SenderType) == chat.ChatSenderType_CHAT_SENDER_TYPE_USER && session.AgentId == 0 {
		publishQueueEvent(ctx, svcCtx, session)
	}
	return msg, nil
}

func publishMessageEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage) {
	if svcCtx.BusRedis == nil || msg == nil {
		return
	}

	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		CreatedAt: nowMillis(),
		Data:      toProtoMessage(msg),
		Session:   toProtoSession(session),
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat message event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat message event failed: %v", err)
	}
}

func publishQueueEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) {
	if svcCtx.BusRedis == nil || session == nil {
		return
	}
	queue, err := toProtoQueueInfo(ctx, svcCtx, session)
	if err != nil {
		logx.WithContext(ctx).Errorf("build chat queue event failed: %v", err)
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE,
		CreatedAt: nowMillis(),
		Data:      newEventSystemMessage(session, queue.GetMessage()),
		Session:   toProtoSession(session),
		Queue:     queue,
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat queue event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat queue event failed: %v", err)
	}
}

func publishSessionEvent(ctx context.Context, svcCtx *svc.ServiceContext, eventType chat.ChatEventType, session *models.TChatSession, operatorID int64, assignType chat.ChatAssignType, reason, message string) {
	if svcCtx.BusRedis == nil || session == nil {
		return
	}
	queue, err := toProtoQueueInfo(ctx, svcCtx, session)
	if err != nil {
		logx.WithContext(ctx).Errorf("build chat session event queue failed: %v", err)
	}
	sessionEvent := &chat.ChatSessionEvent{
		SessionNo:  session.SessionNo,
		MerchantId: session.MerchantId,
		UserId:     session.UserId,
		AgentId:    session.AgentId,
		OperatorId: operatorID,
		Status:     chat.ChatSessionStatus(session.Status),
		AssignType: assignType,
		Reason:     strings.TrimSpace(reason),
		Message:    strings.TrimSpace(message),
		Session:    toProtoSession(session),
		Queue:      queue,
		CreatedAt:  nowMillis(),
	}
	event := &chat.ChatMessageEvent{
		Type:         eventType,
		CreatedAt:    sessionEvent.CreatedAt,
		Data:         newEventSystemMessage(session, sessionEvent.GetMessage()),
		Session:      sessionEvent.Session,
		SessionEvent: sessionEvent,
		Queue:        queue,
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat session event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat session event failed: %v", err)
	}
}

func publishAgentStatusEvent(ctx context.Context, svcCtx *svc.ServiceContext, agent *models.TChatAgent) {
	if svcCtx.BusRedis == nil || agent == nil {
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN,
		CreatedAt: nowMillis(),
		Agent:     toProtoAgent(agent),
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat agent status event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat agent status event failed: %v", err)
	}
}

func publishEvaluationEvent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, satisfaction *models.TChatSatisfaction) {
	if svcCtx.BusRedis == nil || session == nil || satisfaction == nil {
		return
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		CreatedAt: nowMillis(),
		Data: &chat.ChatMessage{
			MessageNo:   nextNo("GM"),
			SessionNo:   session.SessionNo,
			EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
			MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_EVALUATION,
			Content:     "用户已提交评价",
			Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
			AgentId:     strconv.FormatInt(session.AgentId, 10),
			Extra:       satisfactionPayloadJSON(satisfaction),
			CreateTime:  nowMillis(),
			UpdateTime:  nowMillis(),
			Sender: &chat.ChatMessageUser{
				Id:       session.UserId,
				Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
				Nickname: session.Title,
			},
		},
		Session: toProtoSession(session),
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  session.SessionNo,
			MerchantId: session.MerchantId,
			UserId:     session.UserId,
			AgentId:    session.AgentId,
			Status:     chat.ChatSessionStatus(session.Status),
			Message:    "用户已提交评价",
			Session:    toProtoSession(session),
			CreatedAt:  nowMillis(),
		},
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorf("marshal chat evaluation event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.WithContext(ctx).Errorf("publish chat evaluation event failed: %v", err)
	}
}

func satisfactionPayloadJSON(data *models.TChatSatisfaction) string {
	if data == nil {
		return ""
	}
	payload := map[string]interface{}{
		"score":   data.Score,
		"content": data.Content,
		"tags":    data.Tags,
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func changeAgentSessionCount(ctx context.Context, svcCtx *svc.ServiceContext, agentID int64, delta int64) error {
	if agentID <= 0 || delta == 0 {
		return nil
	}
	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, agentID)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	agent.CurrentSessionCount += delta
	if agent.CurrentSessionCount < 0 {
		agent.CurrentSessionCount = 0
	}
	agent.UpdateTimes = nowMillis()
	return svcCtx.ChatAgentModel.Update(ctx, agent)
}

type assignSessionOptions struct {
	SessionNo  string
	ToAgentId  int64
	AssignType chat.ChatAssignType
	Reason     string
}

func assignSession(ctx context.Context, svcCtx *svc.ServiceContext, in assignSessionOptions) (*models.TChatSession, *common.RespBase, error) {
	merchantID, base, err := merchantIDFromMetadata(ctx)
	if base != nil || err != nil {
		return nil, base, err
	}
	session, base, err := getSession(ctx, svcCtx, merchantID, in.SessionNo)
	if base != nil || err != nil {
		return nil, base, err
	}
	if in.ToAgentId <= 0 {
		return nil, badBase("to_agent_id is required"), nil
	}
	assignType := normalizeAssignType(in.AssignType)
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, badBase("chat session is closed"), nil
	}
	if base := validateAssignableSession(session, assignType); base != nil {
		return nil, base, nil
	}
	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, in.ToAgentId)
	if err == models.ErrNotFound || agent.MerchantId != merchantID {
		return nil, notFoundBase("chat agent not found"), nil
	}
	if err != nil {
		return nil, nil, err
	}

	fromAgentID := session.AgentId
	if fromAgentID != agent.Id {
		if err := changeAgentSessionCount(ctx, svcCtx, fromAgentID, -1); err != nil {
			return nil, nil, err
		}
		if err := changeAgentSessionCount(ctx, svcCtx, agent.Id, 1); err != nil {
			return nil, nil, err
		}
	}

	now := nowMillis()
	session.AgentId = agent.Id
	session.GroupId = agent.GroupId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, nil, err
	}

	_, err = svcCtx.ChatAssignmentModel.Insert(ctx, &models.TChatAssignment{
		SessionNo:   session.SessionNo,
		MerchantId:  session.MerchantId,
		FromAgentId: fromAgentID,
		ToAgentId:   agent.Id,
		AssignType:  int64(assignType),
		Reason:      strings.TrimSpace(in.Reason),
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, nil, err
	}

	return session, nil, nil
}

func validateAssignableSession(session *models.TChatSession, assignType chat.ChatAssignType) *common.RespBase {
	switch assignType {
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_AUTO:
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return badBase("chat session cannot be assigned")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL:
		if session.AgentId != 0 {
			return badBase("chat session is already accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return badBase("chat session cannot be accepted")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_TRANSFER:
		if session.AgentId == 0 {
			return badBase("chat session is not accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return badBase("chat session cannot be transferred")
		}
	}
	return nil
}

func releaseSessionToPool(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) (*models.TChatSession, *common.RespBase, error) {
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, badBase("chat session is closed"), nil
	}

	fromAgentID := session.AgentId
	if fromAgentID > 0 {
		if err := changeAgentSessionCount(ctx, svcCtx, fromAgentID, -1); err != nil {
			return nil, nil, err
		}
	}

	now := nowMillis()
	session.AgentId = 0
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, nil, err
	}

	if fromAgentID > 0 {
		_, err := svcCtx.ChatAssignmentModel.Insert(ctx, &models.TChatAssignment{
			SessionNo:   session.SessionNo,
			MerchantId:  session.MerchantId,
			FromAgentId: fromAgentID,
			ToAgentId:   0,
			AssignType:  int64(chat.ChatAssignType_CHAT_ASSIGN_TYPE_TRANSFER),
			Reason:      strings.TrimSpace(reason),
			CreateTimes: now,
			UpdateTimes: now,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return session, nil, nil
}

func routeSessionToAvailableAgent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) (*models.TChatSession, *common.RespBase, error) {
	agents, err := svcCtx.ChatAgentModel.FindAvailable(ctx, session.MerchantId, session.GroupId, 2)
	if err != nil {
		return nil, nil, err
	}
	if len(agents) == 1 {
		if session.AgentId == agents[0].Id {
			return session, nil, nil
		}
		return assignSession(ctx, svcCtx, assignSessionOptions{
			SessionNo:  session.SessionNo,
			ToAgentId:  agents[0].Id,
			AssignType: chat.ChatAssignType_CHAT_ASSIGN_TYPE_AUTO,
			Reason:     reason,
		})
	}
	if session.AgentId > 0 {
		return releaseSessionToPool(ctx, svcCtx, session, reason)
	}
	return session, nil, nil
}

func prepareSessionForUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*models.TChatSession, *common.RespBase, error) {
	if session.AgentId == 0 || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return routeSessionToAvailableAgent(ctx, svcCtx, session, "auto assign")
	}

	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, session.AgentId)
	if err == nil && agent.MerchantId == session.MerchantId && agent.Status == int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE) {
		return session, nil, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, nil, err
	}

	return routeSessionToAvailableAgent(ctx, svcCtx, session, "current agent unavailable")
}

func autoAssignSession(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) error {
	if session.AgentId != 0 || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil
	}
	_, _, err := routeSessionToAvailableAgent(ctx, svcCtx, session, "auto assign")
	return err
}

func markRead(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, readerType chat.ChatSenderType, readerID int64) error {
	now := nowMillis()
	lastNo := session.LastMessageNo
	cursor, err := svcCtx.ChatReadCursorModel.FindOneByMerchantIdSessionNoReaderTypeReaderIdDeviceId(ctx, session.MerchantId, session.SessionNo, int64(readerType), readerID, defaultDeviceID)
	switch err {
	case models.ErrNotFound:
		_, err = svcCtx.ChatReadCursorModel.Insert(ctx, &models.TChatReadCursor{
			MerchantId:        session.MerchantId,
			SessionNo:         session.SessionNo,
			ReaderType:        int64(readerType),
			ReaderId:          readerID,
			DeviceId:          defaultDeviceID,
			LastReadMessageNo: lastNo,
			LastReadTime:      now,
			CreateTimes:       now,
			UpdateTimes:       now,
		})
	case nil:
		cursor.LastReadMessageNo = lastNo
		cursor.LastReadTime = now
		cursor.UpdateTimes = now
		err = svcCtx.ChatReadCursorModel.Update(ctx, cursor)
	}
	if err != nil {
		return err
	}

	switch readerType {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		session.UserUnreadCount = 0
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		session.AgentUnreadCount = 0
	}
	session.UpdateTimes = now
	return svcCtx.ChatSessionModel.Update(ctx, session)
}

func closeSession(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) error {
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil
	}
	now := nowMillis()
	oldAgentID := session.AgentId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
	session.CloseTime = now
	session.CloseReason = strings.TrimSpace(reason)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return err
	}
	return changeAgentSessionCount(ctx, svcCtx, oldAgentID, -1)
}
