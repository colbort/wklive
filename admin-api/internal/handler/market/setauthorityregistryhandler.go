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

func SetAuthorityRegistryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetAuthorityRegistryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := market.NewSetAuthorityRegistryLogic(r.Context(), svcCtx)
		resp, err := l.SetAuthorityRegistry(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
