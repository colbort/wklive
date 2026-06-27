package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListTransientChatMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListTransientChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListTransientChatMessagesLogic {
	return &AdminListTransientChatMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询游客临时消息
func (l *AdminListTransientChatMessagesLogic) AdminListTransientChatMessages(in *chat.AdminListTransientChatMessagesReq) (*chat.AdminListChatMessagesResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminListChatMessagesResp{}, nil
}
