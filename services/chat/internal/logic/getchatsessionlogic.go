package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatSessionLogic {
	return &GetChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话详情
func (l *GetChatSessionLogic) GetChatSession(in *chat.GetChatSessionReq) (*chat.AdminChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatSessionResp{}, nil
}
