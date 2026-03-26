// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/user_private"
	"wklive/app-api/internal/svc"
)

func GetSecurityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user_private.NewGetSecurityLogic(r.Context(), svcCtx)
		resp, err := l.GetSecurity()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
