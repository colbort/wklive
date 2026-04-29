package itick

import (
	"net/http"

	"wklive/app-api/internal/logic/itick"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetItickOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := itick.NewGetItickOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetItickOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
