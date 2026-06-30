// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_token

import (
	"context"
	"strings"

	"chat-api/internal/jwt"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetChatTokenCookieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetChatTokenCookieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetChatTokenCookieLogic {
	return &SetChatTokenCookieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetChatTokenCookieLogic) SetChatTokenCookie(req *types.SetChatTokenCookieReq) (resp *types.RespBase, token string, expireAt int64, err error) {
	token = strings.TrimSpace(req.ChatToken)
	claims, err := jwt.Verify(l.svcCtx.Config.Jwt.AccessSecret, token)
	if err != nil {
		return &types.RespBase{Code: 400, Msg: err.Error()}, "", 0, nil
	}

	return &types.RespBase{Code: 200, Msg: "ok"}, token, claims.ExpireAt, nil
}
