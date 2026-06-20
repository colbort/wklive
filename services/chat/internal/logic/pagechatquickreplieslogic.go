package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatQuickRepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatQuickRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatQuickRepliesLogic {
	return &PageChatQuickRepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询快捷回复
func (l *PageChatQuickRepliesLogic) PageChatQuickReplies(in *chat.PageChatQuickRepliesReq) (*chat.PageChatQuickRepliesResp, error) {
	// todo: add your logic here and delete this line

	return &chat.PageChatQuickRepliesResp{}, nil
}
