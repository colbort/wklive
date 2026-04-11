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

type AppListAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListAccountsLogic {
	return &AppListAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListAccountsLogic) AppListAccounts(req *types.AppListAccountsReq) (resp *types.AppListAccountsResp, err error) {
	return logicutil.Proxy[types.AppListAccountsResp](l.ctx, req, l.svcCtx.OptionCli.AppListAccounts)
}
