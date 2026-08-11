package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/generate"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

const chatMessageSlowStageThreshold = 100 * time.Millisecond

type SendMessageOptions struct {
	MerchantId     int64
	SessionNo      string
	IsGuest        bool
	Sender         *chat.ChatMessageUser
	Receiver       *chat.ChatMessageUser
	MessageType    chat.ChatMessageType
	Content        string
	Url            string
	FileName       string
	MimeType       string
	FileSize       int64
	Duration       int32
	ReceiveChannel string
	ReceiptChannel string
}

func MessageNextCursor(list []*models.ChatMessage) int64 {
	if len(list) == 0 {
		return 0
	}
	return list[len(list)-1].CreateTimes
}

func SendMessage(ctx context.Context, svcCtx *svc.ServiceContext, opts SendMessageOptions) (*chat.ChatMessage, error) {
	totalStarted := time.Now()
	defer func() {
		logChatMessageStage(ctx, "total", totalStarted, opts.MerchantId, opts.SessionNo, "", nil)
	}()

	stageStarted := time.Now()
	var session *models.TChatSession
	var err error
	session, err = GetSession(ctx, svcCtx, opts.MerchantId, opts.SessionNo, opts.IsGuest)
	logChatMessageStage(ctx, "get_session", stageStarted, opts.MerchantId, opts.SessionNo, "", err)
	if err != nil {
		return nil, err
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, errors.New("chat session is closed")
	}
	stageStarted = time.Now()
	mmg, err := buildMessage(ctx, svcCtx, session, opts)
	logChatMessageStage(ctx, "build_message", stageStarted, opts.MerchantId, opts.SessionNo, "", err)
	if err != nil {
		return nil, err
	}
	var msg *chat.ChatMessage
	if opts.IsGuest {
		// 游客/临时会话
		stageStarted = time.Now()
		msg, err = AppendTransientMessage(ctx, svcCtx.Redis, opts.MerchantId, ToProtoMessage(mmg), session)
		logChatMessageStage(ctx, "append_transient_message", stageStarted, opts.MerchantId, opts.SessionNo, mmg.MessageNo, err)
		if err != nil {
			return nil, err
		}
	} else {
		// 非游客
		if opts.Sender == nil {
			return nil, errors.New("sender is required")
		}
		stageStarted = time.Now()
		mmg, err = sendPersistedMessage(ctx, svcCtx, session, mmg)
		messageNo := ""
		if mmg != nil {
			messageNo = mmg.MessageNo
		}
		logChatMessageStage(ctx, "persist_message", stageStarted, opts.MerchantId, opts.SessionNo, messageNo, err)
		if err != nil {
			return nil, err
		}
		msg = ToProtoMessage(mmg)
	}

	stageStarted = time.Now()
	err = PublishMessageEvent(ctx, svcCtx.MQPublisher, opts.ReceiveChannel, PublishEventMessage, &chat.ChatWsResponse_Message{Message: msg})
	logChatMessageStage(ctx, "publish_message_event", stageStarted, opts.MerchantId, opts.SessionNo, msg.MessageNo, err)
	if err != nil {
		return nil, err
	}
	stageStarted = time.Now()
	err = PublishMessageEvent(ctx, svcCtx.MQPublisher, opts.ReceiptChannel, PublishEventMessageDelivered, &chat.ChatWsResponse_Receipt{Receipt: &chat.ChatMessageReceiptPayload{
		SessionNo:     msg.SessionNo,
		MessageNo:     msg.MessageNo,
		SenderId:      msg.Sender.Id,
		OperatorId:    msg.Receiver.Id,
		OperatorType:  chat.ChatSenderType(msg.Receiver.Type),
		MessageStatus: chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_DELIVERED,
		ReceiptTime:   utils.NowMillis(),
	}})
	logChatMessageStage(ctx, "publish_message_receipt", stageStarted, opts.MerchantId, opts.SessionNo, msg.MessageNo, err)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func logChatMessageStage(ctx context.Context, stage string, started time.Time, merchantID int64, sessionNo, messageNo string, err error) {
	duration := time.Since(started)
	if err == nil && duration < chatMessageSlowStageThreshold {
		return
	}
	logx.WithContext(ctx).WithDuration(duration).Slowf(
		"[CHAT_MESSAGE] stage=%s merchant_id=%d session_no=%s message_no=%s duration_ms=%d err=%v",
		stage, merchantID, sessionNo, messageNo, duration.Milliseconds(), err,
	)
}

func buildMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, opts SendMessageOptions) (*models.ChatMessage, error) {
	stageStarted := time.Now()
	messageNo, err := generate.GenerateNo(svcCtx.Redis, ctx, "chat", "CM", "")
	logChatMessageStage(ctx, "generate_message_no", stageStarted, opts.MerchantId, opts.SessionNo, messageNo, err)
	if err != nil {
		return nil, errorx.Wrapf(err, "generate message no error")
	}
	stageStarted = time.Now()
	chatUser, err := svcCtx.ChatUserModel.FindOne(ctx, session.AgentUserId)
	logChatMessageStage(ctx, "find_chat_user", stageStarted, opts.MerchantId, opts.SessionNo, messageNo, err)
	if err != nil {
		return nil, errorx.Wrapf(err, "chat user err: chat user id is %d", session.AgentUserId)
	}
	if chatUser == nil {
		return nil, errors.New("chat user not found")
	}

	now := utils.NowMillis()
	message := models.ChatMessage{
		MessageNo:   messageNo,
		SessionNo:   session.SessionNo,
		MerchantId:  session.MerchantId,
		MessageType: int64(opts.MessageType),
		Content:     opts.Content,
		Url:         opts.Url,
		FileName:    opts.FileName,
		MimeType:    opts.MimeType,
		FileSize:    opts.FileSize,
		Duration:    opts.Duration,
		Status:      int64(chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT),
		CreateTimes: now,
		UpdateTimes: now,
	}
	if opts.Sender != nil {
		message.Sender = &models.ChatMessageUser{
			Id:        opts.Sender.GetId(),
			Type:      int64(opts.Sender.Type),
			Nickname:  opts.Sender.GetNickname(),
			AvatarUrl: opts.Sender.GetAvatarUrl(),
		}
	}
	if opts.Receiver == nil && opts.Sender.GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		opts.Receiver = SessionMessageUser(session)
	}
	if opts.Receiver != nil {
		message.Receiver = &models.ChatMessageUser{
			Id:        opts.Receiver.GetId(),
			Type:      int64(opts.Receiver.Type),
			Nickname:  opts.Receiver.GetNickname(),
			AvatarUrl: opts.Receiver.GetAvatarUrl(),
		}
	}
	return &message, nil
}

func SessionMessageUser(session *models.TChatSession) *chat.ChatMessageUser {
	if session == nil {
		return nil
	}
	user := &chat.ChatMessageUser{
		Id:   session.UserId,
		Type: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
	}
	var ext map[string]string
	if session.ExtJson.Valid && strings.TrimSpace(session.ExtJson.String) != "" {
		if err := json.Unmarshal([]byte(session.ExtJson.String), &ext); err == nil {
			user.Nickname = strings.TrimSpace(ext["nickname"])
			user.AvatarUrl = strings.TrimSpace(ext["avatarUrl"])
		}
	}
	if user.Nickname == "" {
		user.Nickname = strings.TrimSpace(session.Title)
	}
	return user
}

func sendPersistedMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage) (*models.ChatMessage, error) {
	model := svcCtx.ChatMessageFactory.New(session.MerchantId)
	if model == nil {
		return nil, fmt.Errorf("invalid merchant_id: %d", session.MerchantId)
	}
	stageStarted := time.Now()
	if err := model.Insert(ctx, msg); err != nil {
		logChatMessageStage(ctx, "mongo_insert_message", stageStarted, session.MerchantId, session.SessionNo, msg.MessageNo, err)
		return nil, err
	}
	logChatMessageStage(ctx, "mongo_insert_message", stageStarted, session.MerchantId, session.SessionNo, msg.MessageNo, nil)

	now := msg.CreateTimes
	session.LastMessageNo = msg.MessageNo
	session.LastMessage = TrimSummary(msg.Content, msg.FileName, msg.Url)
	session.LastSenderType = int64(msg.Sender.Type)
	session.LastMessageTime = now
	session.UpdateTimes = now
	switch chat.ChatSenderType(msg.Sender.Type) {
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
	stageStarted = time.Now()
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		logChatMessageStage(ctx, "mysql_update_session", stageStarted, session.MerchantId, session.SessionNo, msg.MessageNo, err)
		return nil, err
	}
	logChatMessageStage(ctx, "mysql_update_session", stageStarted, session.MerchantId, session.SessionNo, msg.MessageNo, nil)

	return msg, nil
}
