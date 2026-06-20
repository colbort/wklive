package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
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
	if in.GetId() <= 0 || in.GetMerchantId() <= 0 {
		return &chat.AdminChatAgentResp{Base: badBase("id and merchant_id are required")}, nil
	}
	data, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != in.GetMerchantId() {
		return &chat.AdminChatAgentResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	if in.GetMaxSessionCount() > 0 {
		data.MaxSessionCount = int64(in.GetMaxSessionCount())
	}
	data.GroupId = in.GetGroupId()
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = nowMillis()
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatAgentResp{Base: okBase(), Data: toProtoAgent(data)}, nil
}
