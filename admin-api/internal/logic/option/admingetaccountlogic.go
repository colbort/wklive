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

type AdminGetAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetAccountLogic {
	return &AdminGetAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetAccountLogic) AdminGetAccount(req *types.GetAccountReq) (resp *types.GetAccountResp, err error) {
	return logicutil.Proxy[types.GetAccountResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetAccount)
}
