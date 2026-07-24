package liquiditylogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportQuoteOrderStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportQuoteOrderStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportQuoteOrderStateLogic {
	return &ReportQuoteOrderStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReportQuoteOrderStateLogic) ReportQuoteOrderState(in *liquidity.ReportQuoteOrderStateReq) (*liquidity.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.CommonResp{}, nil
}
