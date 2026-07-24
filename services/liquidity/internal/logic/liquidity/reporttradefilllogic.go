package liquiditylogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportTradeFillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportTradeFillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportTradeFillLogic {
	return &ReportTradeFillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReportTradeFillLogic) ReportTradeFill(in *liquidity.ReportTradeFillReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
