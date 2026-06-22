package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatSessionLogic {
	return &GetChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话详情
func (l *GetChatSessionLogic) GetChatSession(in *chat.GetChatSessionReq) (*chat.AdminChatSessionResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	data, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AdminChatSessionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	return &chat.AdminChatSessionResp{Base: okBase(), Data: toProtoSession(data)}, nil
}
