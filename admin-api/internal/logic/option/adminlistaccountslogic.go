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

type AdminListAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListAccountsLogic {
	return &AdminListAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListAccountsLogic) AdminListAccounts(req *types.ListAccountsReq) (resp *types.ListAccountsResp, err error) {
	return logicutil.Proxy[types.ListAccountsResp](l.ctx, req, l.svcCtx.OptionCli.AdminListAccounts)
}
