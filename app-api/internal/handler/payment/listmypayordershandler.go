// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/payment"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

func ListMyPayOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMyPayOrdersReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := payment.NewListMyPayOrdersLogic(r.Context(), svcCtx)
		resp, err := l.ListMyPayOrders(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
