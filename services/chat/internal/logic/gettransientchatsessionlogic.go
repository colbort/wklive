package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTransientChatSessionLogic {
	return &GetTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询游客临时会话
func (l *GetTransientChatSessionLogic) GetTransientChatSession(in *chat.GetTransientChatSessionReq) (*chat.AppChatSessionResp, error) {
	session, err := internal.GetTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetSessionNo())
	if err != nil {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(404, "transient chat session not found")}, nil
	}
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: session}, nil
}
