package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatAgentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatAgentStatusLogic {
	return &UpdateChatAgentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新坐席在线状态
func (l *UpdateChatAgentStatusLogic) UpdateChatAgentStatus(in *chat.UpdateChatAgentStatusReq) (*chat.AdminChatAgentResp, error) {
	if in.Id <= 0 || in.GetStatus() == chat.ChatAgentStatus_CHAT_AGENT_STATUS_UNKNOWN {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(400, "agent_id and status are required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatAgentResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(404, "chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data.Status = int64(in.GetStatus())
	data.LastActiveTime = utils.NowMillis()
	data.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	_ = ih.PublishMessageEvent(l.ctx, l.svcCtx, ih.PublishMessageEventReq{
		EventType:    chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE,
		Channel:      chat.ChatAdminEventChannel,
		Agent:        data,
		EventMessage: "坐席状态已更新",
	})
	return &chat.AdminChatAgentResp{Base: helper.OkResp(), Data: ih.ToProtoAgent(data)}, nil
}
