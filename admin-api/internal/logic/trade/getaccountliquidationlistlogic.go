// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountLiquidationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidationListLogic {
	return &GetAccountLiquidationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountLiquidationListLogic) GetAccountLiquidationList(req *types.GetAccountLiquidationListReq) (resp *types.GetAccountLiquidationListResp, err error) {
	return logicutil.Proxy[types.GetAccountLiquidationListResp](l.ctx, req, l.svcCtx.TradeCli.GetAccountLiquidationList)
}
