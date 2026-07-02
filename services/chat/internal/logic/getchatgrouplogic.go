package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatGroupModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(404, "chat group not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatGroupResp{Base: helper.OkResp(), Data: ih.ToProtoChatGroup(data)}, nil
}
