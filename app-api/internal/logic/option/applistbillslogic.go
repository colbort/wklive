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

type AppListBillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListBillsLogic {
	return &AppListBillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListBillsLogic) AppListBills(req *types.AppListBillsReq) (resp *types.AppListBillsResp, err error) {
	return logicutil.Proxy[types.AppListBillsResp](l.ctx, req, l.svcCtx.OptionCli.AppListBills)
}
