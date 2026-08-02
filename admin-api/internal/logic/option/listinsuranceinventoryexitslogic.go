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

type ListInsuranceInventoryExitsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListInsuranceInventoryExitsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInsuranceInventoryExitsLogic {
	return &ListInsuranceInventoryExitsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListInsuranceInventoryExitsLogic) ListInsuranceInventoryExits(req *types.ListInsuranceInventoryExitsReq) (resp *types.ListInsuranceInventoryExitsResp, err error) {
	return logicutil.Proxy[types.ListInsuranceInventoryExitsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListInsuranceInventoryExits,
	)
}
