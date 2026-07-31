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

type HaltContractTradingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHaltContractTradingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HaltContractTradingLogic {
	return &HaltContractTradingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HaltContractTradingLogic) HaltContractTrading(req *types.HaltContractTradingReq) (resp *types.GetTradingHaltResp, err error) {
	return logicutil.Proxy[types.GetTradingHaltResp](
		l.ctx, req, l.svcCtx.OptionCli.HaltContractTrading,
	)
}
