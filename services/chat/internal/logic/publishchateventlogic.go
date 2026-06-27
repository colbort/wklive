package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishChatEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishChatEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishChatEventLogic {
	return &PublishChatEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布客服消息事件
func (l *PublishChatEventLogic) PublishChatEvent(in *chat.PublishChatEventReq) (*chat.PublishChatEventResp, error) {
	if err := internal.PublishChatEvent(l.ctx, l.svcCtx, in.GetEvent()); err != nil {
		return &chat.PublishChatEventResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.PublishChatEventResp{Base: helper.OkResp()}, nil
}
