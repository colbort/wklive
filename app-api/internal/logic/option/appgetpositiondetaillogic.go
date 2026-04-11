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

type AppGetPositionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetPositionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetPositionDetailLogic {
	return &AppGetPositionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppGetPositionDetailLogic) AppGetPositionDetail(req *types.AppGetPositionDetailReq) (resp *types.AppGetPositionDetailResp, err error) {
	return logicutil.Proxy[types.AppGetPositionDetailResp](l.ctx, req, l.svcCtx.OptionCli.AppGetPositionDetail)
}
