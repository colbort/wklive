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

type CreateSettlementPriceCorrectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSettlementPriceCorrectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSettlementPriceCorrectionLogic {
	return &CreateSettlementPriceCorrectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSettlementPriceCorrectionLogic) CreateSettlementPriceCorrection(req *types.CreateSettlementPriceCorrectionReq) (resp *types.GetSettlementPriceResp, err error) {
	return logicutil.Proxy[types.GetSettlementPriceResp](
		l.ctx, req, l.svcCtx.OptionCli.CreateSettlementPriceCorrection,
	)
}
