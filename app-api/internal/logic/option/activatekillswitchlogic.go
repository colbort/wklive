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

type ActivateKillSwitchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 激活用户期权 kill switch
func NewActivateKillSwitchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivateKillSwitchLogic {
	return &ActivateKillSwitchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivateKillSwitchLogic) ActivateKillSwitch(req *types.ActivateKillSwitchReq) (resp *types.GetUserTradingControlResp, err error) {
	return logicutil.Proxy[types.GetUserTradingControlResp](
		l.ctx, req, l.svcCtx.OptionCli.ActivateKillSwitch,
	)
}
