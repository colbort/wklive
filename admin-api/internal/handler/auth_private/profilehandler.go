// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"net/http"

	"wklive/admin-api/internal/logic/auth_private"
	"wklive/admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth_private.NewProfileLogic(r.Context(), svcCtx)
		resp, err := l.Profile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
