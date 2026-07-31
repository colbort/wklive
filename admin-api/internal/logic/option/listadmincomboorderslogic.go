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

type ListAdminComboOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAdminComboOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminComboOrdersLogic {
	return &ListAdminComboOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAdminComboOrdersLogic) ListAdminComboOrders(req *types.ListAdminComboOrdersReq) (resp *types.ListAdminComboOrdersResp, err error) {
	return logicutil.Proxy[types.ListAdminComboOrdersResp](
		l.ctx, req, l.svcCtx.OptionCli.ListAdminComboOrders,
	)
}
