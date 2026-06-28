// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_token

import (
	"net/http"
	"time"

	"chat-api/internal/jwt"
	"chat-api/internal/logic/chat_token"
	"chat-api/internal/svc"
	"chat-api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateChatTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateChatTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat_token.NewCreateChatTokenLogic(r.Context(), svcCtx)
		resp, err := l.CreateChatToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if resp != nil && resp.Code == 200 {
				setChatTokenCookie(w, resp.Data.ChatToken, resp.Data.ExpireAt)
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func setChatTokenCookie(w http.ResponseWriter, token string, expireAt int64) {
	if token == "" {
		return
	}

	expires := time.UnixMilli(expireAt)
	maxAge := int(time.Until(expires).Seconds())
	if maxAge <= 0 {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     jwt.ChatTokenCookieName,
		Value:    token,
		Path:     "/chat",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}
