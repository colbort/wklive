package logic

import (
	"context"

	"wklive/proto/chat"
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
	if in.GetMerchantId() <= 0 || in.GetId() <= 0 {
		return &chat.AdminChatAgentResp{Base: badBase("merchant_id and id are required")}, nil
	}
	data, err := l.svcCtx.ChatAgentModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != in.GetMerchantId() {
		return &chat.AdminChatAgentResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatAgentResp{Base: okBase(), Data: toProtoAgent(data)}, nil
}
