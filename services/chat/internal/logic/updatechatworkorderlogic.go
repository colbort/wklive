package logic

import (
	"context"
	"database/sql"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatWorkOrderLogic {
	return &UpdateChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新工单
func (l *UpdateChatWorkOrderLogic) UpdateChatWorkOrder(in *chat.UpdateChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	title := strings.TrimSpace(in.GetTitle())
	content := strings.TrimSpace(in.GetContent())
	if title == "" || content == "" {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "title and content are required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatWorkOrderResp{Base: base}, nil
	}
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
	if status <= 0 {
		status = data.Status
	}
	data.AgentId = in.GetAgentId()
	data.GroupId = in.GetGroupId()
	data.Title = title
	data.Content = sql.NullString{String: content, Valid: true}
	data.ContactName = strings.TrimSpace(in.GetContactName())
	data.ContactMobile = strings.TrimSpace(in.GetContactMobile())
	data.ContactEmail = strings.TrimSpace(in.GetContactEmail())
	data.Priority = int64(in.GetPriority())
	data.Status = status
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = utils.NowMillis()
	if (in.Status == 3 || in.Status == 4) && data.FinishTime == 0 {
		data.FinishTime = data.UpdateTimes
	}
	if err := l.svcCtx.ChatWorkOrderModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatWorkOrderResp{Base: helper.OkResp(), Data: ih.ToProtoChatWorkOrder(data)}, nil
}
