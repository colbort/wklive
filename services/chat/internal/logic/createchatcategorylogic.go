package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatCategoryLogic {
	return &CreateChatCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建问题分类
func (l *CreateChatCategoryLogic) CreateChatCategory(in *chat.CreateChatCategoryReq) (*chat.AdminChatCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatCategoryResp{}, nil
}
