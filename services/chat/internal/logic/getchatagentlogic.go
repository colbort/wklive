package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatAgentLogic {
	return &GetChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询坐席详情
func (l *GetChatAgentLogic) GetChatAgent(in *chat.GetChatAgentReq) (*chat.AdminChatAgentResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatAgentResp{}, nil
}
