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

type OpenChatSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenChatSessionLogic {
	return &OpenChatSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpenChatSessionLogic) OpenChatSession(req *types.OpenChatSessionReq) (resp *types.ChatSessionResp, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.OpenChatSessionReq
		UserId int64
	}{
		OpenChatSessionReq: req,
		UserId:             userId,
	}
	return logicutil.Proxy[types.ChatSessionResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.OpenChatSession)
}
