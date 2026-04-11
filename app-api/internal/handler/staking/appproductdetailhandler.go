// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/staking"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

func AppProductDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AppProductDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := staking.NewAppProductDetailLogic(r.Context(), svcCtx)
		resp, err := l.AppProductDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
