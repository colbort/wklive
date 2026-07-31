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

type GetComboOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组合订单及不可变腿
func NewGetComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComboOrderLogic {
	return &GetComboOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetComboOrderLogic) GetComboOrder(req *types.OptionGetComboOrderReq) (resp *types.OptionGetComboOrderResp, err error) {
	return logicutil.Proxy[types.OptionGetComboOrderResp](
		l.ctx, req, l.svcCtx.OptionCli.GetComboOrder,
	)
}
