package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatWorkOrderLogic {
	return &DeleteChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除工单
func (l *DeleteChatWorkOrderLogic) DeleteChatWorkOrder(in *chat.DeleteChatWorkOrderReq) (*chat.AdminCommonResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminCommonResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, base, err := internal.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminCommonResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatWorkOrderModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	if err := l.svcCtx.ChatWorkOrderModel.Delete(l.ctx, in.GetId()); err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}
