package logic

import (
	"context"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatAgentLogic {
	return &UpdateChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新坐席
func (l *UpdateChatAgentLogic) UpdateChatAgent(in *chat.UpdateChatAgentReq) (*chat.AdminChatAgentResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(400, "id is required")}, nil
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
	if in.GetMaxSessionCount() > 0 {
		data.MaxSessionCount = int64(in.GetMaxSessionCount())
	}
	if in.GetAutoOnline() != 0 {
		data.AutoOnline = int64(in.GetAutoOnline())
	}
	if in.WelcomeMessage != "" {
		data.WelcomeMessage = strings.TrimSpace(in.GetWelcomeMessage())
	}
	if in.GroupId > 0 {
		data.GroupId = in.GetGroupId()
	}
	if strings.TrimSpace(in.GetRemark()) != "" {
		data.Remark = strings.TrimSpace(in.GetRemark())
	}

	data.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatAgentResp{Base: helper.OkResp(), Data: internal.ToProtoAgent(data)}, nil
}
