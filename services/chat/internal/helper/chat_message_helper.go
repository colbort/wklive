package helper

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

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
	MessageChannel string
	ReceiptChannel string
}

func MessageNextCursor(list []*models.ChatMessage) int64 {
	if len(list) == 0 {
		return 0
	}
	return list[len(list)-1].CreateTimes
}

func SendMessage(ctx context.Context, svcCtx *svc.ServiceContext, opts SendMessageOptions) (*chat.ChatMessage, *common.RespBase, error) {
	var session *models.TChatSession
	var base *common.RespBase
	var err error
	session, base, err = GetSession(ctx, svcCtx, opts.MerchantId, opts.SessionNo, opts.IsGuest)
	if err != nil || base != nil {
		return nil, base, err
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, helper.ErrResp(400, "chat session is closed"), nil
	}
	mmg, base, err := buildMessage(ctx, svcCtx, session, opts)
	if err != nil || base != nil {
		return nil, base, err
	}
	var msg *chat.ChatMessage
	if opts.IsGuest {
		// 游客/临时会话
		msg, err = AppendTransientMessage(ctx, svcCtx.BusRedis, opts.MerchantId, ToProtoMessage(mmg), session)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// 非游客
		if opts.Sender == nil {
			return nil, helper.ErrResp(400, "sender is required"), nil
		}
		mmg, err = sendPersistedMessage(ctx, svcCtx, session, mmg)
		if err != nil {
			return nil, nil, err
		}
		msg = ToProtoMessage(mmg)
	}

	_ = PublishMessageEvent(PublishMessageEventReq{
		Ctx:       ctx,
		BusRedis:  svcCtx.BusRedis,
		Channel:   opts.MessageChannel,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		Payload:   chat.ChatMessageEvent_Message{Message: msg},
	})
	_ = PublishMessageEvent(PublishMessageEventReq{
		Ctx:       ctx,
		BusRedis:  svcCtx.BusRedis,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED,
		Channel:   opts.ReceiptChannel,
		Payload: chat.ChatMessageEvent_Receipt{Receipt: &chat.ChatMessageReceiptPayload{
			SessionNo:     msg.SessionNo,
			MessageNo:     msg.MessageNo,
			SenderId:      msg.Sender.Id,
			OperatorId:    msg.Receiver.Id,
			OperatorType:  chat.ChatSenderType(msg.Receiver.Type),
			MessageStatus: chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_DELIVERED,
			ReceiptTime:   utils.NowMillis(),
		}},
	})
	return nil, helper.ErrResp(400, "sender type is invalid"), nil
}

func buildMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, opts SendMessageOptions) (*models.ChatMessage, *common.RespBase, error) {
	messageNo, err := svcCtx.GenerateNo(ctx, "CM")
	if err != nil {
		return nil, helper.ErrResp(400, "generate message no error"), nil
	}
	chatUser, err := svcCtx.ChatUserModel.FindOne(ctx, session.AgentUserId)
	if err != nil {
		return nil, helper.ErrResp(400, "chat user err"), nil
	}
	if chatUser == nil {
		return nil, helper.ErrResp(400, "chat user not found"), nil
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
	if opts.Receiver != nil {
		message.Receiver = &models.ChatMessageUser{
			Id:        opts.Receiver.GetId(),
			Type:      int64(opts.Receiver.Type),
			Nickname:  opts.Receiver.GetNickname(),
			AvatarUrl: opts.Receiver.GetAvatarUrl(),
		}
	}
	// 用户端发送的消息，接收方是坐席端；Receiver 在这里赋值
	if message.Sender.Type == int64(chat.ChatSenderType_CHAT_SENDER_TYPE_USER) {
		message.Receiver = &models.ChatMessageUser{
			Id:        chatUser.Id,
			Type:      int64(chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT),
			Nickname:  chatUser.Nickname,
			AvatarUrl: chatUser.AvatarUrl,
		}
	} else {
		// 坐席端发送的消息，接收方是用户端；Receiver 信息在发送新的时候已经赋值；
		message.Sender = &models.ChatMessageUser{
			Nickname:  chatUser.Nickname,
			AvatarUrl: chatUser.AvatarUrl,
		}
	}
	return &message, nil, nil
}

func sendPersistedMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, msg *models.ChatMessage) (*models.ChatMessage, error) {
	model := svcCtx.ChatMessageFactory.New(session.MerchantId)
	if model == nil {
		return nil, fmt.Errorf("invalid merchant_id: %d", session.MerchantId)
	}
	if err := model.Insert(ctx, msg); err != nil {
		return nil, err
	}

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
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, err
	}

	return msg, nil
}
