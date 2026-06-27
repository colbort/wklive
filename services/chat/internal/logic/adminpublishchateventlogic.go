package logic

import (
	"context"

	"wklive/proto/chat"
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
	// todo: add your logic here and delete this line

	return &chat.AdminPublishChatEventResp{}, nil
}
