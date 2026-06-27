package logic

import (
	"context"
	"fmt"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAgentMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendAgentMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAgentMessageLogic {
	return &SendAgentMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送客服消息
func (l *SendAgentMessageLogic) SendAgentMessage(in *chat.SendAgentMessageReq) (*chat.AdminChatMessageResp, error) {
	if in.GetIsGuest() || in.GetMessage() != nil || in.GetPublishOnly() {
		msg := in.GetMessage()
		var err error
		if !in.GetPublishOnly() {
			msg, err = internal.AppendTransientMessage(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), msg, in.GetSession(), in.GetTtlSeconds())
			if err != nil {
				return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		if err := internal.PublishTransientMessageEvent(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetEventType(), msg, in.GetSession(), chat.ChatAppMessageChannel); err != nil {
			return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		return &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg}, nil
	}

	session, base, err := internal.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.SessionNo)
	if err != nil {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.AdminChatMessageResp{Base: base}, nil
	}
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, in.MerchantId, in.Sender.Id)
	if err == models.ErrNotFound {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(404, "chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if session.AgentId != 0 && session.AgentId != agent.Id {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(400, "agent does not own this session")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(400, "chat session is closed")}, nil
	}
	if session.AgentId == 0 {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(400, "chat session is not accepted")}, nil
	}
	messageNo, err := l.svcCtx.GenerateNo(l.ctx, "CM")
	if err != nil {
		logx.Errorf("generate message no error: %v", err)
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(400, "generate message no error")}, nil
	}
	now := utils.NowMillis()
	msg, err := internal.SendMessage(l.ctx, l.svcCtx, session, &models.ChatMessage{
		MessageNo:  messageNo,
		SessionNo:  session.SessionNo,
		MerchantId: session.MerchantId,
		Sender: &models.ChatMessageUser{
			Id:        in.Sender.Id,
			Type:      int64(in.Sender.Type),
			Nickname:  in.Sender.Nickname,
			AvatarUrl: in.Sender.AvatarUrl,
		},
		Receiver: &models.ChatMessageUser{
			Id:        session.UserId,
			Type:      int64(chat.ChatSenderType_CHAT_SENDER_TYPE_USER),
			Nickname:  fmt.Sprintf("User%d", session.Id),
			AvatarUrl: "",
		},
		MessageType: int64(in.MessageType),
		Content:     in.Content,
		Url:         in.Url,
		FileName:    in.FileName,
		MimeType:    in.MimeType,
		FileSize:    in.FileSize,
		Duration:    in.Duration,
		Status:      int64(chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT),
		CreateTimes: now,
		UpdateTimes: now,
	}, chat.ChatAppMessageChannel)
	if err != nil {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: internal.ToProtoMessage(msg)}, nil
}
