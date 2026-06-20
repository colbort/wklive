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
	session, base, err := getSession(l.ctx, l.svcCtx, in.GetMerchantId(), in.GetSessionNo())
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if session.UserId != in.GetUserId() {
		return &chat.AppChatSessionResp{Base: notFoundBase("chat session not found")}, nil
	}
	if err := closeSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
