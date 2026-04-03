// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/itick"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func GetProductKlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetProductKlineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := itick.NewGetProductKlineLogic(r.Context(), svcCtx)
		resp, err := l.GetProductKline(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
