package logic

import (
	"context"

	"wklive/proto/chat"
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
	if in.GetAgentId() <= 0 || in.GetStatus() == chat.ChatAgentStatus_CHAT_AGENT_STATUS_UNKNOWN {
		return &chat.AdminChatAgentResp{Base: badBase("agent_id and status are required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatAgentResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.GetAgentId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatAgentResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	data.Status = int64(in.GetStatus())
	data.LastActiveTime = nowMillis()
	data.UpdateTimes = data.LastActiveTime
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatAgentResp{Base: okBase(), Data: toProtoAgent(data)}, nil
}
