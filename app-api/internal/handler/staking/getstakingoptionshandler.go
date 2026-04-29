package staking

import (
	"net/http"

	"wklive/app-api/internal/logic/staking"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetStakingOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := staking.NewGetStakingOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetStakingOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
