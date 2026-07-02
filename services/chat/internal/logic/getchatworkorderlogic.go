package logic

import (
	"context"
	"strings"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatWorkOrderLogic {
	return &GetChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询工单详情
func (l *GetChatWorkOrderLogic) GetChatWorkOrder(in *chat.GetChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	if in.GetId() <= 0 && strings.TrimSpace(in.GetWorkOrderNo()) == "" {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "id or work_order_no is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	var data *models.TChatWorkOrder
	if in.GetId() > 0 {
		data, err = l.svcCtx.ChatWorkOrderModel.FindOne(l.ctx, in.GetId())
	} else {
		data, err = l.svcCtx.ChatWorkOrderModel.FindOneByWorkOrderNo(l.ctx, strings.TrimSpace(in.GetWorkOrderNo()))
	}
	if err == models.ErrNotFound {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	return &chat.AdminChatWorkOrderResp{Base: helper.OkResp(), Data: ih.ToProtoChatWorkOrder(data)}, nil
}
