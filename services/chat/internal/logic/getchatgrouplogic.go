package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatGroupLogic {
	return &GetChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询客服分组详情
func (l *GetChatGroupLogic) GetChatGroup(in *chat.GetChatGroupReq) (*chat.AdminChatGroupResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatGroupResp{Base: badBase("id is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatGroupResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatGroupModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatGroupResp{Base: notFoundBase("chat group not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatGroupResp{Base: okBase(), Data: toProtoChatGroup(data)}, nil
}
