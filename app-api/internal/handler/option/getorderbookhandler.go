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

// 查询来自 Option 活动限价委托的聚合盘口
func GetOrderBookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrderBookReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := option.NewGetOrderBookLogic(r.Context(), svcCtx)
		resp, err := l.GetOrderBook(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
