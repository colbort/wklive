package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkUserMessagesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkUserMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkUserMessagesReadLogic {
	return &MarkUserMessagesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记用户侧已读
func (l *MarkUserMessagesReadLogic) MarkUserMessagesRead(in *chat.MarkUserMessagesReadReq) (*chat.AppMarkMessagesReadResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AppMarkMessagesReadResp{}, nil
}
