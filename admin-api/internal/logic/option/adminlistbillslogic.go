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

type AdminListBillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListBillsLogic {
	return &AdminListBillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListBillsLogic) AdminListBills(req *types.ListBillsReq) (resp *types.ListBillsResp, err error) {
	return logicutil.Proxy[types.ListBillsResp](l.ctx, req, l.svcCtx.OptionCli.AdminListBills)
}
