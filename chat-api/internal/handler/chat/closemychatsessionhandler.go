// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"net/http"

	"chat-api/internal/jwt"
	"chat-api/internal/logic/chat"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CloseMyChatSessionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CloseMyChatSessionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		claims, err := jwt.Verify(svcCtx.Config.Jwt.AccessSecret, jwt.TokenFromRequest(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		ctx := jwt.ContextWithClaims(r.Context(), claims)
		ctx = contextWithChatIdentity(ctx, claims.MerchantId, claims.UserId)

		l := chat.NewCloseMyChatSessionLogic(ctx, svcCtx)
		resp, err := l.CloseMyChatSession(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
