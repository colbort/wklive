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

type AdminGetUserTradingControlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetUserTradingControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetUserTradingControlLogic {
	return &AdminGetUserTradingControlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetUserTradingControlLogic) AdminGetUserTradingControl(req *types.AdminGetUserTradingControlReq) (resp *types.AdminGetUserTradingControlResp, err error) {
	return logicutil.Proxy[types.AdminGetUserTradingControlResp](
		l.ctx, req, l.svcCtx.OptionCli.AdminGetUserTradingControl,
	)
}
