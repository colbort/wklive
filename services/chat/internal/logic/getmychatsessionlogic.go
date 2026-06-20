package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyChatSessionLogic {
	return &GetMyChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话详情
func (l *GetMyChatSessionLogic) GetMyChatSession(in *chat.GetMyChatSessionReq) (*chat.AppChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AppChatSessionResp{}, nil
}
