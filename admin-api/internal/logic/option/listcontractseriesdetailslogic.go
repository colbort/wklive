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

type ListContractSeriesDetailsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListContractSeriesDetailsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContractSeriesDetailsLogic {
	return &ListContractSeriesDetailsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListContractSeriesDetailsLogic) ListContractSeriesDetails(req *types.ListContractSeriesDetailsReq) (resp *types.ListContractSeriesDetailsResp, err error) {
	return logicutil.Proxy[types.ListContractSeriesDetailsResp](l.ctx, req, l.svcCtx.OptionCli.ListContractSeriesDetails)
}
