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

type DeleteChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatGroupLogic {
	return &DeleteChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除客服分组
func (l *DeleteChatGroupLogic) DeleteChatGroup(in *chat.DeleteChatGroupReq) (*chat.AdminCommonResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminCommonResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatGroupModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat group not found")}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat group not found")}, nil
	}
	if err := l.svcCtx.ChatGroupModel.Delete(l.ctx, in.GetId()); err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}
