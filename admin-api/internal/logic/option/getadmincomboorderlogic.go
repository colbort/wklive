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

type GetAdminComboOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAdminComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminComboOrderLogic {
	return &GetAdminComboOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAdminComboOrderLogic) GetAdminComboOrder(req *types.GetAdminComboOrderReq) (resp *types.GetAdminComboOrderResp, err error) {
	return logicutil.Proxy[types.GetAdminComboOrderResp](
		l.ctx, req, l.svcCtx.OptionCli.GetAdminComboOrder,
	)
}
