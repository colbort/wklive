package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatAgentLogic {
	return &UpdateChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新坐席
func (l *UpdateChatAgentLogic) UpdateChatAgent(in *chat.UpdateChatAgentReq) (*chat.AdminChatAgentResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatAgentResp{}, nil
}
