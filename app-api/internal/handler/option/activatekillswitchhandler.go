// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/option"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

// 激活用户期权 kill switch
func ActivateKillSwitchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ActivateKillSwitchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := option.NewActivateKillSwitchLogic(r.Context(), svcCtx)
		resp, err := l.ActivateKillSwitch(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
