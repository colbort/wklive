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

type ListContractSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContractSeriesLogic {
	return &ListContractSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListContractSeriesLogic) ListContractSeries(req *types.ListContractSeriesReq) (resp *types.ListContractSeriesResp, err error) {
	return logicutil.Proxy[types.ListContractSeriesResp](l.ctx, req, l.svcCtx.OptionCli.ListContractSeries)
}
