package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatCategoriesLogic {
	return &PageChatCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询问题分类
func (l *PageChatCategoriesLogic) PageChatCategories(in *chat.PageChatCategoriesReq) (*chat.PageChatCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &chat.PageChatCategoriesResp{}, nil
}
