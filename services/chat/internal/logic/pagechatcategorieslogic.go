package logic

import (
	"context"
	"strings"
	"wklive/common/helper"
	"wklive/common/pageutil"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
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
	merchantID, base, err := internal.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatCategoriesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatCategoriesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	list, total, err := l.svcCtx.ChatCategoryModel.FindPage(l.ctx, models.ChatCategoryPageFilter{
		MerchantId:   merchantID,
		ParentId:     in.GetParentId(),
		CategoryCode: strings.TrimSpace(in.GetCategoryCode()),
		CategoryName: strings.TrimSpace(in.GetCategoryName()),
		Enabled:      int64(in.GetEnabled()),
		Keyword:      strings.TrimSpace(in.GetKeyword()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatCategoriesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.PageChatCategoriesResp{
		Base: internal.OffsetBase(cursor, limit, len(list), total),
		Data: internal.ToProtoChatCategories(list),
	}, nil
}
