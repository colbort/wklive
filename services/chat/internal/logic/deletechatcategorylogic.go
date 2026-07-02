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
	if in.GetId() <= 0 {
		return &chat.AdminCommonResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatCategoryModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat category not found")}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat category not found")}, nil
	}
	if err := l.svcCtx.ChatCategoryModel.Delete(l.ctx, in.GetId()); err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}
