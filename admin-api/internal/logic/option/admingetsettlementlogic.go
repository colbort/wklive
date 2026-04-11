// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetSettlementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetSettlementLogic {
	return &AdminGetSettlementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetSettlementLogic) AdminGetSettlement(req *types.GetSettlementReq) (resp *types.GetSettlementResp, err error) {
	return logicutil.Proxy[types.GetSettlementResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetSettlement)
}
