package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEnabledChatQuickRepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEnabledChatQuickRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEnabledChatQuickRepliesLogic {
	return &ListEnabledChatQuickRepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询启用快捷回复
func (l *ListEnabledChatQuickRepliesLogic) ListEnabledChatQuickReplies(in *chat.ListEnabledChatQuickRepliesReq) (*chat.ListChatQuickRepliesResp, error) {
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.ListChatQuickRepliesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	list, err := l.svcCtx.ChatQuickReplyModel.ListEnabled(l.ctx, merchantID, in.GetAgentId(), in.GetCategoryId())
	if err != nil {
		return &chat.ListChatQuickRepliesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.ListChatQuickRepliesResp{Base: helper.OkResp(), Data: ih.ToProtoChatQuickReplies(list)}, nil
}
