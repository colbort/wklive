// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/option"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func RetrySettlementInstructionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OptionRetrySettlementInstructionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := option.NewRetrySettlementInstructionLogic(r.Context(), svcCtx)
		resp, err := l.RetrySettlementInstruction(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
