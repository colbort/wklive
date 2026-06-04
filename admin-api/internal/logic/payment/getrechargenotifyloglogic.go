// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeNotifyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRechargeNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeNotifyLogLogic {
	return &GetRechargeNotifyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRechargeNotifyLogLogic) GetRechargeNotifyLog(req *types.GetRechargeNotifyLogReq) (resp *types.GetRechargeNotifyLogResp, err error) {
	return logicutil.Proxy[types.GetRechargeNotifyLogResp](l.ctx, req, l.svcCtx.PaymentCli.GetRechargeNotifyLog)
}
