package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatCategoryLogic {
	return &GetChatCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询问题分类详情
func (l *GetChatCategoryLogic) GetChatCategory(in *chat.GetChatCategoryReq) (*chat.AdminChatCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatCategoryResp{}, nil
}
