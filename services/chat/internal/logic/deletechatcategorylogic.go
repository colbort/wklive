package logic

import (
	"context"

	"wklive/proto/chat"
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
		return &chat.AdminCommonResp{Base: badBase("id is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminCommonResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatCategoryModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminCommonResp{Base: notFoundBase("chat category not found")}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: errorBase(err)}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminCommonResp{Base: notFoundBase("chat category not found")}, nil
	}
	if err := l.svcCtx.ChatCategoryModel.Delete(l.ctx, in.GetId()); err != nil {
		return &chat.AdminCommonResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminCommonResp{Base: okBase()}, nil
}
