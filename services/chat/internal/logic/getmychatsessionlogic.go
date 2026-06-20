package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyChatSessionLogic {
	return &GetMyChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话详情
func (l *GetMyChatSessionLogic) GetMyChatSession(in *chat.GetMyChatSessionReq) (*chat.AppChatSessionResp, error) {
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
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
}
