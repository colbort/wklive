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

type MarkUserMessagesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkUserMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkUserMessagesReadLogic {
	return &MarkUserMessagesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkUserMessagesReadLogic) MarkUserMessagesRead(req *types.MarkUserMessagesReadReq) (resp *types.RespBase, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.MarkUserMessagesReadReq
		UserId int64
	}{
		MarkUserMessagesReadReq: req,
		UserId:                  userId,
	}
	return logicutil.Proxy[types.RespBase](l.ctx, proxyReq, l.svcCtx.ChatAppCli.MarkUserMessagesRead)
}
