package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppUpsertTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppUpsertTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppUpsertTransientChatSessionLogic {
	return &AppUpsertTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或更新游客临时会话
func (l *AppUpsertTransientChatSessionLogic) AppUpsertTransientChatSession(in *chat.AppUpsertTransientChatSessionReq) (*chat.AppChatSessionResp, error) {
	session, err := internal.UpsertTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetSession(), in.GetTtlSeconds())
	if err != nil {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: session}, nil
}
