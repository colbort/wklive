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

func ChangePriceFormulaStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangePriceFormulaStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := itick.NewChangePriceFormulaStatusLogic(r.Context(), svcCtx)
		resp, err := l.ChangePriceFormulaStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
