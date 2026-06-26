package logic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/proto/chat"
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
		return &chat.AdminChatWorkOrderResp{Base: badBase("title and content are required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatWorkOrderResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatWorkOrderResp{Base: errorBase(err)}, nil
	}

	workOrderNo := ""
	exists := true
	for attempt := 0; attempt < sessionNoInsertAttempts; attempt++ {
		workOrderNo = nextNo("WO")
		_, err := l.svcCtx.ChatWorkOrderModel.FindOneByWorkOrderNo(l.ctx, workOrderNo)
		if err == models.ErrNotFound {
			exists = false
			break
		}
	}
	if workOrderNo == "" || exists {
		return &chat.AdminChatWorkOrderResp{Base: helper.FailResp()}, nil
	}
	now := nowMillis()
	data := &models.TChatWorkOrder{
		MerchantId:    merchantID,
		WorkOrderNo:   nextNo("WO"),
		SessionNo:     strings.TrimSpace(in.GetSessionNo()),
		UserId:        in.GetUserId(),
		AgentId:       in.GetAgentId(),
		GroupId:       in.GetGroupId(),
		Title:         title,
		Content:       nullString(content),
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
		return &chat.AdminChatWorkOrderResp{Base: okBase(), Data: toProtoChatWorkOrder(data)}, nil
	} else {
		return &chat.AdminChatWorkOrderResp{Base: errorBase(err)}, nil
	}
}
