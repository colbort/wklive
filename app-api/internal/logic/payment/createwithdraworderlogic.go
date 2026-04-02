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

type CreateWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawOrderLogic {
	return &CreateWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateWithdrawOrderLogic) CreateWithdrawOrder(req *types.CreateWithdrawOrderReq) (resp *types.CreateWithdrawOrderResp, err error) {
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

	result, err := l.svcCtx.PaymentCli.CreateWithdrawOrder(l.ctx, &payment.CreateWithdrawOrderReq{
		TenantId: tenantId,
		UserId:   userId,
		Amount:   req.Amount,
		Currency: req.Currency,
		Address:  req.Address,
		BankId:   req.BankId,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.CreateWithdrawOrderResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Id: result.Id,
	}

	return
}
