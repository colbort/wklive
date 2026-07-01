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

type CreateChatWorkOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatWorkOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatWorkOrderLogic {
	return &CreateChatWorkOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建工单/离线留言
func (l *CreateChatWorkOrderLogic) CreateChatWorkOrder(in *chat.CreateChatWorkOrderReq) (*chat.AdminChatWorkOrderResp, error) {
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
	workOrderNo, err := l.svcCtx.GenerateNo(l.ctx, "WO")
	if err != nil {
		logx.Errorf("generate work order no error: %v", err)
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(400, "generate message no error")}, nil
	}
	now := utils.NowMillis()
	data := &models.TChatWorkOrder{
		MerchantId:    merchantID,
		WorkOrderNo:   workOrderNo,
		SessionNo:     strings.TrimSpace(in.GetSessionNo()),
		UserId:        in.GetUserId(),
		AgentId:       in.GetAgentId(),
		GroupId:       in.GetGroupId(),
		Title:         title,
		Content:       sql.NullString{String: content, Valid: true},
		ContactName:   strings.TrimSpace(in.GetContactName()),
		ContactMobile: strings.TrimSpace(in.GetContactMobile()),
		ContactEmail:  strings.TrimSpace(in.GetContactEmail()),
		Priority:      int64(in.Priority),
		Status:        1,
		Remark:        strings.TrimSpace(in.GetRemark()),
		CreateTimes:   now,
		UpdateTimes:   now,
	}
	result, err := l.svcCtx.ChatWorkOrderModel.Insert(l.ctx, data)
	if err == nil {
		if id, err := result.LastInsertId(); err == nil {
			data.Id = id
		}
		return &chat.AdminChatWorkOrderResp{Base: helper.OkResp(), Data: ih.ToProtoChatWorkOrder(data)}, nil
	} else {
		return &chat.AdminChatWorkOrderResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
}
