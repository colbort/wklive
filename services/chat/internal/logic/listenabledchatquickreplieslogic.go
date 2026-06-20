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
	// todo: add your logic here and delete this line

	return &chat.ListChatQuickRepliesResp{}, nil
}
