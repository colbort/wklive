package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTransientChatSessionLogic {
	return &DeleteTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除游客临时会话和消息
func (l *DeleteTransientChatSessionLogic) DeleteTransientChatSession(in *chat.DeleteTransientChatSessionReq) (*chat.DeleteTransientChatSessionResp, error) {
	if err := internal.DeleteTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetSessionNo()); err != nil {
		return &chat.DeleteTransientChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.DeleteTransientChatSessionResp{Base: helper.OkResp()}, nil
}
