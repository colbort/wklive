// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"net/http"

	"chat-admin-api/internal/logic/chat_auth"
	"chat-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func OptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := chat_auth.NewOptionsLogic(r.Context(), svcCtx)
		resp, err := l.Options()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
