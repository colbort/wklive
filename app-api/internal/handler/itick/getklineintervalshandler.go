// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/itick"
	"wklive/app-api/internal/svc"
)

func GetKlineIntervalsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := itick.NewGetKlineIntervalsLogic(r.Context(), svcCtx)
		resp, err := l.GetKlineIntervals()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
