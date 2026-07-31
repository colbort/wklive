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

type ListSettlementPricesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSettlementPricesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettlementPricesLogic {
	return &ListSettlementPricesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSettlementPricesLogic) ListSettlementPrices(req *types.ListSettlementPricesReq) (resp *types.ListSettlementPricesResp, err error) {
	return logicutil.Proxy[types.ListSettlementPricesResp](
		l.ctx, req, l.svcCtx.OptionCli.ListSettlementPrices,
	)
}
