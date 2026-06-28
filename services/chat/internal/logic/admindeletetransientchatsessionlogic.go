package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteTransientChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDeleteTransientChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteTransientChatSessionLogic {
	return &AdminDeleteTransientChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除游客临时会话和消息
func (l *AdminDeleteTransientChatSessionLogic) AdminDeleteTransientChatSession(in *chat.AdminDeleteTransientChatSessionReq) (*chat.AdminDeleteTransientChatSessionResp, error) {
	session, _ := internal.GetTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetSessionNo())
	if err := internal.PublishMessageEvent(l.ctx, l.svcCtx, internal.PublishMessageEventReq{
		EventType:        in.GetEventType(),
		Channel:          chat.ChatAppMessageChannel,
		MerchantId:       in.GetMerchantId(),
		SessionNo:        in.GetSessionNo(),
		UserId:           in.GetUserId(),
		AgentId:          in.GetAgentId(),
		EventMessage:     in.GetEventMessage(),
		TransientSession: session,
	}); err != nil {
		return &chat.AdminDeleteTransientChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if err := internal.DeleteTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetSessionNo()); err != nil {
		return &chat.AdminDeleteTransientChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminDeleteTransientChatSessionResp{Base: helper.OkResp()}, nil
}
