package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatQuickReplyLogic {
	return &CreateChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建快捷回复
func (l *CreateChatQuickReplyLogic) CreateChatQuickReply(in *chat.CreateChatQuickReplyReq) (*chat.AdminChatQuickReplyResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatQuickReplyResp{}, nil
}
