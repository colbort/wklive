package logic

import (
	"context"

	"wklive/proto/chat"
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
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.ListChatQuickRepliesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.ListChatQuickRepliesResp{Base: errorBase(err)}, nil
	}
	list, err := l.svcCtx.ChatQuickReplyModel.ListEnabled(l.ctx, merchantID, in.GetAgentId(), in.GetCategoryId())
	if err != nil {
		return &chat.ListChatQuickRepliesResp{Base: errorBase(err)}, nil
	}
	return &chat.ListChatQuickRepliesResp{Base: okBase(), Data: toProtoChatQuickReplies(list)}, nil
}
