package logic

import (
	"context"

	"wklive/proto/chat"
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
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	if err := closeSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
