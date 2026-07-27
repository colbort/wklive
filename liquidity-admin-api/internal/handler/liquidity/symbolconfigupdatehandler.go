package liquidity

import (
	"net/http"

	"wklive/liquidity-admin-api/internal/logic/liquidity"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SymbolConfigUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveSymbolConfigReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := liquidity.NewSymbolConfigUpdateLogic(r.Context(), svcCtx)
		resp, err := l.SymbolConfigUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
