// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package market

import (
	"net/http"

	"wklive/admin-api/internal/logic/market"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreatePriceFormulaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreatePriceFormulaReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := market.NewCreatePriceFormulaLogic(r.Context(), svcCtx)
		resp, err := l.CreatePriceFormula(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
