// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"chat-api/internal/logicutil"
	"wklive/common/utils"

	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatSessionsLogic {
	return &ListMyChatSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyChatSessionsLogic) ListMyChatSessions(req *types.ListChatSessionsReq) (resp *types.ListChatSessionsResp, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.ListChatSessionsReq
		UserId int64
	}{
		ListChatSessionsReq: req,
		UserId:              userId,
	}
	return logicutil.Proxy[types.ListChatSessionsResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.ListMyChatSessions)
}
