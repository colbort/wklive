// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"context"

	"chat-api/internal/logicutil"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthChatMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthChatMerchantLogic {
	return &AuthChatMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthChatMerchantLogic) AuthChatMerchant(req *types.ChatAuthReq) (resp *types.ChatAuthResp, err error) {
	return logicutil.Proxy[types.ChatAuthResp](l.ctx, req, l.svcCtx.ChatAppCli.AuthChatMerchant)
}
