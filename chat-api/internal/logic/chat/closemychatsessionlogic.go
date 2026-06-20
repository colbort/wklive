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

type CloseMyChatSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCloseMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseMyChatSessionLogic {
	return &CloseMyChatSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseMyChatSessionLogic) CloseMyChatSession(req *types.CloseMyChatSessionReq) (resp *types.ChatSessionResp, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.CloseMyChatSessionReq
		UserId int64
	}{
		CloseMyChatSessionReq: req,
		UserId:                userId,
	}
	return logicutil.Proxy[types.ChatSessionResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.CloseMyChatSession)
}
