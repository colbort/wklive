package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkAgentMessagesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAgentMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAgentMessagesReadLogic {
	return &MarkAgentMessagesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记客服侧已读
func (l *MarkAgentMessagesReadLogic) MarkAgentMessagesRead(in *chat.MarkAgentMessagesReadReq) (*chat.AdminMarkMessagesReadResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminMarkMessagesReadResp{}, nil
}
