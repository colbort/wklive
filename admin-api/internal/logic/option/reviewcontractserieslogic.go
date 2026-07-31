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

type ReviewContractSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewContractSeriesLogic {
	return &ReviewContractSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewContractSeriesLogic) ReviewContractSeries(req *types.ReviewContractSeriesReq) (resp *types.GetContractSeriesResp, err error) {
	return logicutil.Proxy[types.GetContractSeriesResp](l.ctx, req, l.svcCtx.OptionCli.ReviewContractSeries)
}
