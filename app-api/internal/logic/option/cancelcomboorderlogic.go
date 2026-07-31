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

type CancelComboOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 原子撤销组合父单并释放所有未完成腿冻结
func NewCancelComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelComboOrderLogic {
	return &CancelComboOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelComboOrderLogic) CancelComboOrder(req *types.OptionCancelComboOrderReq) (resp *types.CommonResp, err error) {
	return logicutil.Proxy[types.CommonResp](l.ctx, req, l.svcCtx.OptionCli.CancelComboOrder)
}
