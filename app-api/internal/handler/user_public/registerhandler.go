// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"net/http"

	"wklive/app-api/internal/logic/user_public"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user_public.NewRegisterLogic(r.Context(), svcCtx)
		req.RegisterIp = utils.GetClientIP(r)
		resp, err := l.Register(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
