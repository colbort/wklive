// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_token

import (
	"net/http"

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
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
