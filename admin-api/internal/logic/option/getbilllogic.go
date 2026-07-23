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

type GetBillLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBillLogic {
	return &GetBillLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBillLogic) GetBill(req *types.GetBillReq) (resp *types.GetBillResp, err error) {
	return logicutil.Proxy[types.GetBillResp](l.ctx, req, l.svcCtx.OptionCli.GetBill)
}
