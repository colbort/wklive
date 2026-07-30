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

type ListLiquidationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLiquidationsLogic {
	return &ListLiquidationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLiquidationsLogic) ListLiquidations(req *types.ListOptionLiquidationsReq) (resp *types.ListOptionLiquidationsResp, err error) {
	return logicutil.Proxy[types.ListOptionLiquidationsResp](l.ctx, req, l.svcCtx.OptionCli.ListLiquidations)
}
