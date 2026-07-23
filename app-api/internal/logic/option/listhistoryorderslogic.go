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

type ListHistoryOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListHistoryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListHistoryOrdersLogic {
	return &ListHistoryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListHistoryOrdersLogic) ListHistoryOrders(req *types.ListHistoryOrdersReq) (resp *types.ListHistoryOrdersResp, err error) {
	return logicutil.Proxy[types.ListHistoryOrdersResp](l.ctx, req, l.svcCtx.OptionCli.ListHistoryOrders)
}
