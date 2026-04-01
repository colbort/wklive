// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/itick"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

func GetKlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetKlineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := itick.NewGetKlineLogic(r.Context(), svcCtx)
		resp, err := l.GetKline(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
