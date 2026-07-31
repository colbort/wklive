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

type GetOperationsOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOperationsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperationsOverviewLogic {
	return &GetOperationsOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOperationsOverviewLogic) GetOperationsOverview(req *types.GetOperationsOverviewReq) (resp *types.GetOperationsOverviewResp, err error) {
	return logicutil.Proxy[types.GetOperationsOverviewResp](
		l.ctx, req, l.svcCtx.OptionCli.GetOperationsOverview,
	)
}
