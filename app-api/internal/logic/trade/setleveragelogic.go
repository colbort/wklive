// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetLeverageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetLeverageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetLeverageLogic {
	return &SetLeverageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetLeverageLogic) SetLeverage(req *types.SetLeverageReq) (resp *types.TradeAppCommonResp, err error) {
	return logicutil.Proxy[types.TradeAppCommonResp](l.ctx, req, l.svcCtx.TradeCli.SetLeverage)
}
