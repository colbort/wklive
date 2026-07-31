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

type ReviewSettlementPriceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewSettlementPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewSettlementPriceLogic {
	return &ReviewSettlementPriceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewSettlementPriceLogic) ReviewSettlementPrice(req *types.ReviewSettlementPriceReq) (resp *types.GetSettlementPriceResp, err error) {
	return logicutil.Proxy[types.GetSettlementPriceResp](
		l.ctx, req, l.svcCtx.OptionCli.ReviewSettlementPrice,
	)
}
