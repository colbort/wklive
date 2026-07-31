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

type ListTradingHaltsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTradingHaltsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradingHaltsLogic {
	return &ListTradingHaltsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTradingHaltsLogic) ListTradingHalts(req *types.ListTradingHaltsReq) (resp *types.ListTradingHaltsResp, err error) {
	return logicutil.Proxy[types.ListTradingHaltsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListTradingHalts,
	)
}
