package logic

import (
	"context"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleChatWorkOrderLogic {
	return &HandleChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理工单
func (l *HandleChatWorkOrderLogic) HandleChatWorkOrder(in *chat.HandleChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	if in.GetStatus() <= 0 {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "status is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatWorkOrderModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(404, "chat work order not found")}, nil
	}
	status := int64(in.GetStatus())
	now := utils.NowMillis()
	data.HandlerId = in.GetHandlerId()
	data.Status = status
	data.HandleResult = strings.TrimSpace(in.GetHandleResult())
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = now
	if in.Status == 3 || in.Status == 4 {
		data.FinishTime = now
	}
	if err := l.svcCtx.ChatWorkOrderModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatWorkOrderResp{Base: helper.OkResp(), Data: ih.ToProtoChatWorkOrder(data)}, nil
}
