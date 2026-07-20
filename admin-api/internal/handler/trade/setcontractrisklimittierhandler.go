// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/trade"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func SetContractRiskLimitTierHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetContractRiskLimitTierReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := trade.NewSetContractRiskLimitTierLogic(r.Context(), svcCtx)
		resp, err := l.SetContractRiskLimitTier(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
