// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/liquidity-admin-api/internal/logic/liquidity"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
)

func ResolveReconcileDifferenceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveReconcileDifferenceReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := liquidity.NewResolveReconcileDifferenceLogic(r.Context(), svcCtx)
		resp, err := l.ResolveReconcileDifference(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
