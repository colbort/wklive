package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminAppendTransientChatMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminAppendTransientChatMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAppendTransientChatMessageLogic {
	return &AdminAppendTransientChatMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 追加游客临时消息并更新会话摘要
func (l *AdminAppendTransientChatMessageLogic) AdminAppendTransientChatMessage(in *chat.AdminAppendTransientChatMessageReq) (*chat.AdminChatMessageResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatMessageResp{}, nil
}
