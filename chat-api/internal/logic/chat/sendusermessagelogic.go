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

type SendUserMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendUserMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendUserMessageLogic {
	return &SendUserMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendUserMessageLogic) SendUserMessage(req *types.SendUserMessageReq) (resp *types.ChatMessageResp, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.SendUserMessageReq
		UserId int64
	}{
		SendUserMessageReq: req,
		UserId:             userId,
	}
	return logicutil.Proxy[types.ChatMessageResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.SendUserMessage)
}
