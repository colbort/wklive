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
	session, created, err := ensureOpenSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetUserId(), normalizeSource(in.GetSource()), in.GetTitle(), in.GetCategory(), chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL, nil)
	if err != nil {
		return &chat.AppChatSessionResp{Base: badBase(err.Error())}, nil
	}
	if created {
		if err := autoAssignSession(l.ctx, l.svcCtx, session); err != nil {
			return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
		}
		if refreshed, err := l.svcCtx.ChatSessionModel.FindOneBySessionNo(l.ctx, session.SessionNo); err == nil {
			session = refreshed
		}
		if strings.TrimSpace(in.GetFirstMessage()) != "" {
			msg := newMessage(session, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, in.GetUserId(), in.GetSenderNickname(), in.GetSenderAvatarUrl(), chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, in.GetFirstMessage(), "", "", "", 0, nil)
			if _, err := sendMessage(l.ctx, l.svcCtx, session, msg); err != nil {
				return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
			}
		}
	}
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
