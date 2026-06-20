package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatAgentLogic {
	return &CreateChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建坐席
func (l *CreateChatAgentLogic) CreateChatAgent(in *chat.CreateChatAgentReq) (*chat.AdminChatAgentResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatAgentResp{}, nil
}
