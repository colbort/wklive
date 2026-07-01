package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

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
	if in.GetId() <= 0 {
		return &chat.AdminChatCategoryResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatCategoryResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatCategoryModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatCategoryResp{Base: helper.ErrResp(404, "chat category not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatCategoryResp{Base: helper.ErrResp(404, "chat category not found")}, nil
	}
	return &chat.AdminChatCategoryResp{Base: helper.OkResp(), Data: ih.ToProtoChatCategory(data)}, nil
}
