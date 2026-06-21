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
	session, base, err := getSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetSessionNo())
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	if session.UserId != in.GetUserId() {
		return &chat.AppChatMessageResp{Base: notFoundBase("chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AppChatMessageResp{Base: badBase("chat session is closed")}, nil
	}
	session, base, err = prepareSessionForUserMessage(l.ctx, l.svcCtx, session)
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, in.GetUserId(), in.GetSenderNickname(), in.GetSenderAvatarUrl(), in.GetMessageType(), in.GetContent(), in.GetMediaUrl(), in.GetMediaName(), in.GetMediaMime(), in.GetMediaSize(), nil)
	msg, err = sendMessage(l.ctx, l.svcCtx, session, msg)
	if err != nil {
		return &chat.AppChatMessageResp{Base: errorBase(err)}, nil
	}
	return &chat.AppChatMessageResp{Base: okBase(), Data: toProtoMessage(msg)}, nil
}
