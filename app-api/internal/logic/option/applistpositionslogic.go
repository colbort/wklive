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

type AppListPositionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListPositionsLogic {
	return &AppListPositionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListPositionsLogic) AppListPositions(req *types.AppListPositionsReq) (resp *types.AppListPositionsResp, err error) {
	return logicutil.Proxy[types.AppListPositionsResp](l.ctx, req, l.svcCtx.OptionCli.AppListPositions)
}
