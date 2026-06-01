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

func GuestLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GuestLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		ctx := utils.ContextWithClientIP(r.Context(), utils.GetClientIP(r))
		l := user_public.NewGuestLoginLogic(ctx, svcCtx)
		resp, err := l.GuestLogin(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
