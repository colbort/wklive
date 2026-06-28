package chat_token

import (
	"net/http"
	"strings"

	"chat-api/internal/jwt"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type setChatTokenCookieReq struct {
	ChatToken string `json:"chatToken"`
}

func SetChatTokenCookieHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setChatTokenCookieReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		token := strings.TrimSpace(req.ChatToken)
		claims, err := jwt.Verify(svcCtx.Config.Jwt.AccessSecret, token)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, types.RespBase{Code: 400, Msg: err.Error()})
			return
		}

		setChatTokenCookie(w, token, claims.ExpireAt)
		httpx.OkJsonCtx(r.Context(), w, types.RespBase{Code: 200, Msg: "ok"})
	}
}
