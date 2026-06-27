package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendUserMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendUserMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendUserMessageLogic {
	return &SendUserMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送用户消息
func (l *SendUserMessageLogic) SendUserMessage(in *chat.SendUserMessageReq) (*chat.AppChatMessageResp, error) {
	session, base, err := internal.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.SessionNo)
	if err != nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	if session.UserId != in.Sender.Id {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(400, "chat session is closed")}, nil
	}
	chatUser, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, session.AgentUserId)
	if err != nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(400, "chat user err")}, nil
	}
	if chatUser == nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(400, "chat user not found")}, nil
	}
	messageNo, err := l.svcCtx.GenerateNo(l.ctx, "CM")
	if err != nil {
		logx.Errorf("generate message no error: %v", err)
		return &chat.AppChatMessageResp{Base: helper.ErrResp(400, "generate message no error")}, nil
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
			Id:        chatUser.Id,
			Type:      int64(chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT),
			Nickname:  chatUser.Nickname,
			AvatarUrl: chatUser.AvatarUrl,
		},
		MessageType: int64(in.GetMessageType()),
		Content:     in.Content,
		Url:         in.Url,
		FileName:    in.FileName,
		MimeType:    in.MimeType,
		FileSize:    in.FileSize,
		Status:      int64(chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT),
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppChatMessageResp{Base: helper.OkResp(), Data: internal.ToProtoMessage(msg)}, nil
}
