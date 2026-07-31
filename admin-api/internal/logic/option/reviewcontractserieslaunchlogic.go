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

type ReviewContractSeriesLaunchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewContractSeriesLaunchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewContractSeriesLaunchLogic {
	return &ReviewContractSeriesLaunchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewContractSeriesLaunchLogic) ReviewContractSeriesLaunch(req *types.ReviewContractSeriesLaunchReq) (resp *types.GetContractSeriesResp, err error) {
	return logicutil.Proxy[types.GetContractSeriesResp](
		l.ctx, req, l.svcCtx.OptionCli.ReviewContractSeriesLaunch,
	)
}
