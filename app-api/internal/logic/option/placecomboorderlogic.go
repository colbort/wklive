// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlaceComboOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建2至4腿独立策略簿组合订单
func NewPlaceComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceComboOrderLogic {
	return &PlaceComboOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlaceComboOrderLogic) PlaceComboOrder(req *types.OptionPlaceComboOrderReq) (resp *types.OptionPlaceComboOrderResp, err error) {
	return logicutil.Proxy[types.OptionPlaceComboOrderResp](
		l.ctx, req, l.svcCtx.OptionCli.PlaceComboOrder,
	)
}
