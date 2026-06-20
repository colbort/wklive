package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatQuickReplyLogic {
	return &UpdateChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新快捷回复
func (l *UpdateChatQuickReplyLogic) UpdateChatQuickReply(in *chat.UpdateChatQuickReplyReq) (*chat.AdminChatQuickReplyResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatQuickReplyResp{}, nil
}
