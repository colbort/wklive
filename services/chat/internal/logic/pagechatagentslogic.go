package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatAgentsLogic {
	return &PageChatAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询坐席
func (l *PageChatAgentsLogic) PageChatAgents(in *chat.PageChatAgentsReq) (*chat.PageChatAgentsResp, error) {
	// todo: add your logic here and delete this line

	return &chat.PageChatAgentsResp{}, nil
}
