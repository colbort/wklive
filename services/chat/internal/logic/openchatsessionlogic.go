package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenChatSessionLogic {
	return &OpenChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或获取当前会话
func (l *OpenChatSessionLogic) OpenChatSession(in *chat.OpenChatSessionReq) (*chat.AppChatSessionResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	session, created, err := ensureOpenSession(l.ctx, l.svcCtx, merchantID, userID, normalizeSource(in.GetSource()), in.GetTitle(), in.GetCategory(), chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL, nil)
	if err != nil {
		return &chat.AppChatSessionResp{Base: badBase(err.Error())}, nil
	}
	if created {
		publishQueueEvent(l.ctx, l.svcCtx, session)
		if strings.TrimSpace(in.GetFirstMessage()) != "" {
			msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, userID, in.GetSenderNickname(), in.GetSenderAvatarUrl(), chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, in.GetFirstMessage(), "", "", "", 0, nil)
			if _, err := sendMessage(l.ctx, l.svcCtx, session, msg); err != nil {
				return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
			}
		}
	}
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
