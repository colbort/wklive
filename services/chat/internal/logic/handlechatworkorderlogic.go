package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
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
		return &chat.AdminChatWorkOrderResp{Base: badBase("id is required")}, nil
	}
	if in.GetStatus() <= 0 {
		return &chat.AdminChatWorkOrderResp{Base: badBase("status is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatWorkOrderResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatWorkOrderModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatWorkOrderResp{Base: notFoundBase("chat work order not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: errorBase(err)}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatWorkOrderResp{Base: notFoundBase("chat work order not found")}, nil
	}
	status := int64(in.GetStatus())
	now := nowMillis()
	data.HandlerId = in.GetHandlerId()
	data.Status = status
	data.HandleResult = strings.TrimSpace(in.GetHandleResult())
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = now
	if isWorkOrderFinished(status) {
		data.FinishTime = now
	}
	if err := l.svcCtx.ChatWorkOrderModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatWorkOrderResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatWorkOrderResp{Base: okBase(), Data: toProtoChatWorkOrder(data)}, nil
}
