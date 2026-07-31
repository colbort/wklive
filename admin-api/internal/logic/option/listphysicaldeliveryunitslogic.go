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

type ListPhysicalDeliveryUnitsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPhysicalDeliveryUnitsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPhysicalDeliveryUnitsLogic {
	return &ListPhysicalDeliveryUnitsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPhysicalDeliveryUnitsLogic) ListPhysicalDeliveryUnits(req *types.ListPhysicalDeliveryUnitsReq) (resp *types.ListPhysicalDeliveryUnitsResp, err error) {
	return logicutil.Proxy[types.ListPhysicalDeliveryUnitsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListPhysicalDeliveryUnits,
	)
}
