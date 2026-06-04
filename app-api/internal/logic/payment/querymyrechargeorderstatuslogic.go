// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyRechargeOrderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyRechargeOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyRechargeOrderStatusLogic {
	return &QueryMyRechargeOrderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyRechargeOrderStatusLogic) QueryMyRechargeOrderStatus(req *types.QueryMyRechargeOrderStatusReq) (resp *types.QueryMyRechargeOrderStatusResp, err error) {
	return logicutil.Proxy[types.QueryMyRechargeOrderStatusResp](l.ctx, req, l.svcCtx.PaymentCli.QueryMyRechargeOrderStatus)
}
