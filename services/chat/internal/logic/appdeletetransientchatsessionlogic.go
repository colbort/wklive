package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppDeleteTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppDeleteTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppDeleteTransientChatSessionLogic {
	return &AppDeleteTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除游客临时会话和消息
func (l *AppDeleteTransientChatSessionLogic) AppDeleteTransientChatSession(in *chat.AppDeleteTransientChatSessionReq) (*chat.AppDeleteTransientChatSessionResp, error) {
	if err := internal.DeleteTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetSessionNo()); err != nil {
		return &chat.AppDeleteTransientChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppDeleteTransientChatSessionResp{Base: helper.OkResp()}, nil
}
