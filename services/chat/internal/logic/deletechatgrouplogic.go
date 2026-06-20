package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatGroupLogic {
	return &DeleteChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除客服分组
func (l *DeleteChatGroupLogic) DeleteChatGroup(in *chat.DeleteChatGroupReq) (*chat.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminCommonResp{}, nil
}
