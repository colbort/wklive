package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppPublishChatEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppPublishChatEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPublishChatEventLogic {
	return &AppPublishChatEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布客服消息事件
func (l *AppPublishChatEventLogic) AppPublishChatEvent(in *chat.AppPublishChatEventReq) (*chat.AppPublishChatEventResp, error) {
	if err := internal.PublishChatEvent(l.ctx, l.svcCtx, in.GetEvent()); err != nil {
		return &chat.AppPublishChatEventResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppPublishChatEventResp{Base: helper.OkResp()}, nil
}
