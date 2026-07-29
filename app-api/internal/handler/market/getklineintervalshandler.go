// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package market

import (
	"net/http"

	"wklive/app-api/internal/logic/market"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetKlineIntervalsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := market.NewGetKlineIntervalsLogic(r.Context(), svcCtx)
		resp, err := l.GetKlineIntervals()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
