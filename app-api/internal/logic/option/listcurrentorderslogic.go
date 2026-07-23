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

type ListCurrentOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCurrentOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCurrentOrdersLogic {
	return &ListCurrentOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCurrentOrdersLogic) ListCurrentOrders(req *types.ListCurrentOrdersReq) (resp *types.ListCurrentOrdersResp, err error) {
	return logicutil.Proxy[types.ListCurrentOrdersResp](l.ctx, req, l.svcCtx.OptionCli.ListCurrentOrders)
}
