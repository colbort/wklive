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

type AdminListContractsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListContractsLogic {
	return &AdminListContractsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListContractsLogic) AdminListContracts(req *types.ListContractsReq) (resp *types.ListContractsResp, err error) {
	return logicutil.Proxy[types.ListContractsResp](l.ctx, req, l.svcCtx.OptionCli.AdminListContracts)
}
