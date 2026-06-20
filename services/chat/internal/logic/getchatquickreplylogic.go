package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

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
	// todo: add your logic here and delete this line

	return &chat.AdminChatQuickReplyResp{}, nil
}
