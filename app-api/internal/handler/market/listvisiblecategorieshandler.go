// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package market

import (
	"net/http"

	"wklive/app-api/internal/logic/market"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListVisibleCategoriesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListVisibleCategoriesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := market.NewListVisibleCategoriesLogic(r.Context(), svcCtx)
		resp, err := l.ListVisibleCategories(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
