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

type ListComboOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分页查询当前用户组合订单
func NewListComboOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListComboOrdersLogic {
	return &ListComboOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListComboOrdersLogic) ListComboOrders(req *types.OptionListComboOrdersReq) (resp *types.OptionListComboOrdersResp, err error) {
	return logicutil.Proxy[types.OptionListComboOrdersResp](
		l.ctx, req, l.svcCtx.OptionCli.ListComboOrders,
	)
}
