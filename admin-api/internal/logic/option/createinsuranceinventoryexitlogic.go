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

type CreateInsuranceInventoryExitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateInsuranceInventoryExitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInsuranceInventoryExitLogic {
	return &CreateInsuranceInventoryExitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateInsuranceInventoryExitLogic) CreateInsuranceInventoryExit(req *types.CreateInsuranceInventoryExitReq) (resp *types.GetInsuranceInventoryExitResp, err error) {
	return logicutil.Proxy[types.GetInsuranceInventoryExitResp](
		l.ctx, req, l.svcCtx.OptionCli.CreateInsuranceInventoryExit,
	)
}
