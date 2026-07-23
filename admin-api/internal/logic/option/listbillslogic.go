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

type ListBillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBillsLogic {
	return &ListBillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBillsLogic) ListBills(req *types.ListBillsReq) (resp *types.ListBillsResp, err error) {
	return logicutil.Proxy[types.ListBillsResp](l.ctx, req, l.svcCtx.OptionCli.ListBills)
}
