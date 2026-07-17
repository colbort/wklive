package system

import (
	"net/http"

	"wklive/admin-api/internal/logic/system"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SysTenantDomainGuestMigrationStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SysTenantDomainGuestMigrationStatsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := system.NewSysTenantDomainGuestMigrationStatsLogic(r.Context(), svcCtx)
		resp, err := l.SysTenantDomainGuestMigrationStats(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
