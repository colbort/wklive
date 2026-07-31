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

type CreateContractSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContractSeriesLogic {
	return &CreateContractSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateContractSeriesLogic) CreateContractSeries(req *types.CreateContractSeriesReq) (resp *types.GetContractSeriesResp, err error) {
	return logicutil.Proxy[types.GetContractSeriesResp](l.ctx, req, l.svcCtx.OptionCli.CreateContractSeries)
}
