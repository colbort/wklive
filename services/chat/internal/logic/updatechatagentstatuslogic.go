package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
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
	merchantID, base, err := internal.MerchantIDFromMetadata(l.ctx)
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
	_ = internal.PublishMessageEvent(l.ctx, l.svcCtx, internal.PublishMessageEventReq{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_JOIN,
		Channel:   chat.ChatAdminEventChannel,
		Agent:     data,
	})
	return &chat.AdminChatAgentResp{Base: helper.OkResp(), Data: internal.ToProtoAgent(data)}, nil
}
