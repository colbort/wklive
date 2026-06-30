package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
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
	merchantID, base, err := internal.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	session, base, err := internal.GetSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo(), false)
	if err != nil {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.AdminChatSessionResp{Base: base}, nil
	}
	return &chat.AdminChatSessionResp{Base: helper.OkResp(), Data: internal.ToProtoSession(session, false)}, nil
}
