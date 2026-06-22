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
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.ListChatCategoriesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.ListChatCategoriesResp{Base: errorBase(err)}, nil
	}
	list, err := l.svcCtx.ChatCategoryModel.ListEnabledByMerchant(l.ctx, merchantID)
	if err != nil {
		return &chat.ListChatCategoriesResp{Base: errorBase(err)}, nil
	}
	return &chat.ListChatCategoriesResp{Base: okBase(), Data: toProtoChatCategories(list)}, nil
}
