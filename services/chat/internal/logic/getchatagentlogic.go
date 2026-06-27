package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatAgentLogic {
	return &GetChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询坐席详情
func (l *GetChatAgentLogic) GetChatAgent(in *chat.GetChatAgentReq) (*chat.AdminChatAgentResp, error) {
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
	return &chat.AdminChatAgentResp{Base: helper.OkResp(), Data: internal.ToProtoAgent(data)}, nil
}
