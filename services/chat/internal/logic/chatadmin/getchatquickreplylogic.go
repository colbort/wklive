package chatadminlogic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatQuickReplyLogic {
	return &GetChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询快捷回复详情
func (l *GetChatQuickReplyLogic) GetChatQuickReply(in *chat.GetChatQuickReplyReq) (*chat.ChatQuickReplyResp, error) {
	if in.GetId() <= 0 {
		return &chat.ChatQuickReplyResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.ChatQuickReplyResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatQuickReplyModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.ChatQuickReplyResp{Base: helper.ErrResp(404, "chat quick reply not found")}, nil
	}
	if err != nil {
		return &chat.ChatQuickReplyResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.ChatQuickReplyResp{Base: helper.ErrResp(404, "chat quick reply not found")}, nil
	}
	return &chat.ChatQuickReplyResp{Base: helper.OkResp(), Data: ih.ToProtoChatQuickReply(data)}, nil
}
