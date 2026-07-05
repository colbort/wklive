// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"chat-api/internal/logicutil"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatConfigLogic {
	return &GetChatConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatConfigLogic) GetChatConfig(req *types.GetChatConfigReq) (resp *types.ChatConfigResp, err error) {
	return logicutil.Proxy[types.ChatConfigResp](l.ctx, req, l.svcCtx.ChatAppCli.GetChatConfig)
}
