package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseMyChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseMyChatSessionLogic {
	return &CloseMyChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭我的会话
func (l *CloseMyChatSessionLogic) CloseMyChatSession(in *chat.CloseMyChatSessionReq) (*chat.AppChatSessionResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.AppChatSessionResp{Base: notFoundBase("chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
	}
	if err := closeSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	publishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSED, session, 0, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, in.GetCloseReason(), "本次会话已结束")
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
