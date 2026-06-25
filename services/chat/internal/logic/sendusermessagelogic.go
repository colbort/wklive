package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

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
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.AppChatMessageResp{Base: notFoundBase("chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AppChatMessageResp{Base: badBase("chat session is closed")}, nil
	}
	msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, userID, in.GetSenderNickname(), in.GetSenderAvatarUrl(), in.GetMessageType(), in.GetContent(), in.GetUrl(), in.GetFileName(), in.GetMimeType(), in.GetFileSize(), nil)
	msg, err = sendMessage(l.ctx, l.svcCtx, session, msg)
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	return &chat.AppChatMessageResp{Base: okBase(), Data: toProtoMessage(msg)}, nil
}
