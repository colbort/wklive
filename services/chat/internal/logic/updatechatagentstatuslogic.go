package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatAgentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatAgentStatusLogic {
	return &UpdateChatAgentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新坐席在线状态
func (l *UpdateChatAgentStatusLogic) UpdateChatAgentStatus(in *chat.UpdateChatAgentStatusReq) (*chat.AdminChatAgentResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatAgentResp{}, nil
}
