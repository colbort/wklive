// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
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
	result, err := l.svcCtx.PaymentCli.CreateWithdrawOrder(l.ctx, &payment.CreateWithdrawOrderReq{
		ClientIp: req.ClientIp,
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
