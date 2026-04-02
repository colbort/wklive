// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMyRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelMyRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyRechargeOrderLogic {
	return &CancelMyRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelMyRechargeOrderLogic) CancelMyRechargeOrder(req *types.CancelMyRechargeOrderReq) (resp *types.RespBase, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	tenantId := req.TenantId
	if tenantId == 0 {
		tenantId, err = utils.GetTenantIdFromCtx(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	result, err := l.svcCtx.PaymentCli.CancelMyRechargeOrder(l.ctx, &payment.CancelMyRechargeOrderReq{
		TenantId: tenantId,
		UserId:   userId,
		OrderNo:  req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}

	return
}
