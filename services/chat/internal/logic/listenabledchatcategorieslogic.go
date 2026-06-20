package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEnabledChatCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEnabledChatCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEnabledChatCategoriesLogic {
	return &ListEnabledChatCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询启用问题分类
func (l *ListEnabledChatCategoriesLogic) ListEnabledChatCategories(in *chat.ListEnabledChatCategoriesReq) (*chat.ListChatCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &chat.ListChatCategoriesResp{}, nil
}
