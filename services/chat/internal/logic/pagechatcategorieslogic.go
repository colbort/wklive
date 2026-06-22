package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

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
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatCategoriesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatCategoriesResp{Base: errorBase(err)}, nil
	}
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatCategoryModel.FindPage(l.ctx, models.ChatCategoryPageFilter{
		MerchantId:   merchantID,
		ParentId:     in.GetParentId(),
		CategoryCode: strings.TrimSpace(in.GetCategoryCode()),
		CategoryName: strings.TrimSpace(in.GetCategoryName()),
		Enabled:      int64(in.GetEnabled()),
		Keyword:      strings.TrimSpace(in.GetKeyword()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatCategoriesResp{Base: errorBase(err)}, nil
	}
	return &chat.PageChatCategoriesResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoChatCategories(list),
	}, nil
}
