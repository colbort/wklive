package chat_token

import (
	"net/http"

	"chat-api/internal/logic/chat_token"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetChatTokenCookieHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetChatTokenCookieReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat_token.NewSetChatTokenCookieLogic(r.Context(), svcCtx)
		resp, token, expireAt, err := l.SetChatTokenCookie(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if resp != nil && resp.Code == 200 {
			setChatTokenCookie(w, token, expireAt)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
