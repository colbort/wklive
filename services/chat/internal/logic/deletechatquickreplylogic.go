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

type DeleteChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatQuickReplyLogic {
	return &DeleteChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除快捷回复
func (l *DeleteChatQuickReplyLogic) DeleteChatQuickReply(in *chat.DeleteChatQuickReplyReq) (*chat.AdminCommonResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminCommonResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminCommonResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatQuickReplyModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat quick reply not found")}, nil
	}
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminCommonResp{Base: helper.ErrResp(404, "chat quick reply not found")}, nil
	}
	if err := l.svcCtx.ChatQuickReplyModel.Delete(l.ctx, in.GetId()); err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}
