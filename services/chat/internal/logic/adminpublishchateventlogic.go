package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminPublishChatEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminPublishChatEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminPublishChatEventLogic {
	return &AdminPublishChatEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布客服消息事件
func (l *AdminPublishChatEventLogic) AdminPublishChatEvent(in *chat.AdminPublishChatEventReq) (*chat.AdminPublishChatEventResp, error) {
	if err := internal.PublishChatEvent(l.ctx, l.svcCtx, in.GetEvent(), chat.ChatAdminEventChannel); err != nil {
		return &chat.AdminPublishChatEventResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminPublishChatEventResp{Base: helper.OkResp()}, nil
}
