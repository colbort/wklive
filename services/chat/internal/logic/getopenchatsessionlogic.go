package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOpenChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOpenChatSessionLogic {
	return &GetOpenChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户未关闭会话
func (l *GetOpenChatSessionLogic) GetOpenChatSession(in *chat.GetOpenChatSessionReq) (*chat.InternalChatSessionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.InternalChatSessionResp{}, nil
}
