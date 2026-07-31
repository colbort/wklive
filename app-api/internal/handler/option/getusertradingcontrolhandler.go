// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/option"
	"wklive/app-api/internal/svc"
)

// 查询用户期权交易控制
func GetUserTradingControlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := option.NewGetUserTradingControlLogic(r.Context(), svcCtx)
		resp, err := l.GetUserTradingControl()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
