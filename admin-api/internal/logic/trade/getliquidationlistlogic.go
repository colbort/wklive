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

type GetLiquidationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLiquidationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLiquidationListLogic {
	return &GetLiquidationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLiquidationListLogic) GetLiquidationList(req *types.GetLiquidationListReq) (resp *types.GetLiquidationListResp, err error) {
	return logicutil.Proxy[types.GetLiquidationListResp](l.ctx, req, l.svcCtx.TradeCli.GetLiquidationList)
}
