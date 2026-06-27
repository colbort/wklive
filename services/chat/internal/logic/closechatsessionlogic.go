package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseChatSessionLogic {
	return &CloseChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭会话
func (l *CloseChatSessionLogic) CloseChatSession(in *chat.CloseChatSessionReq) (*chat.AdminChatSessionResp, error) {
	session, err := l.svcCtx.ChatSessionModel.FindOneBySessionNo(l.ctx, in.SessionNo)
	if err != nil {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(400, "chat session is closed")}, nil
	}
	if err := internal.CloseSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	internal.PublishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, session, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, in.GetCloseReason(), "本次会话已结束")
	return &chat.AdminChatSessionResp{Base: helper.OkResp(), Data: internal.ToProtoSession(session)}, nil
}
