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

type AppGetOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetOrderDetailLogic {
	return &AppGetOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppGetOrderDetailLogic) AppGetOrderDetail(req *types.AppGetOrderDetailReq) (resp *types.AppGetOrderDetailResp, err error) {
	return logicutil.Proxy[types.AppGetOrderDetailResp](l.ctx, req, l.svcCtx.OptionCli.AppGetOrderDetail)
}
