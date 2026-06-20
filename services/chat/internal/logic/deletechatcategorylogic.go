package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatCategoryLogic {
	return &DeleteChatCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除问题分类
func (l *DeleteChatCategoryLogic) DeleteChatCategory(in *chat.DeleteChatCategoryReq) (*chat.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminCommonResp{}, nil
}
