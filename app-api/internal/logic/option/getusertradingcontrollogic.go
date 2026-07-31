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

type GetUserTradingControlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询用户期权交易控制
func NewGetUserTradingControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradingControlLogic {
	return &GetUserTradingControlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTradingControlLogic) GetUserTradingControl() (resp *types.GetUserTradingControlResp, err error) {
	return logicutil.Proxy[types.GetUserTradingControlResp](
		l.ctx, struct{}{}, l.svcCtx.OptionCli.GetUserTradingControl,
	)
}
