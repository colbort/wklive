package logic

import (
	"context"

	"wklive/proto/chat"
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
func (l *GetChatQuickReplyLogic) GetChatQuickReply(in *chat.GetChatQuickReplyReq) (*chat.AdminChatQuickReplyResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatQuickReplyResp{Base: badBase("id is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatQuickReplyResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatQuickReplyModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatQuickReplyResp{Base: notFoundBase("chat quick reply not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatQuickReplyResp{Base: notFoundBase("chat quick reply not found")}, nil
	}
	return &chat.AdminChatQuickReplyResp{Base: okBase(), Data: toProtoChatQuickReply(data)}, nil
}
