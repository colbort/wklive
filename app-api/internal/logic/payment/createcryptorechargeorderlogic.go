// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCryptoRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCryptoRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCryptoRechargeOrderLogic {
	return &CreateCryptoRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCryptoRechargeOrderLogic) CreateCryptoRechargeOrder(req *types.CreateCryptoRechargeOrderReq) (*types.CreateCryptoRechargeOrderResp, error) {
	result, err := l.svcCtx.PaymentCli.CreateCryptoRechargeOrder(l.ctx, &payment.CreateCryptoRechargeOrderReq{
		WalletType:     req.WalletType,
		Coin:           req.Coin,
		ChainCode:      common.ChainCode(req.ChainCode),
		RechargeAmount: req.RechargeAmount,
		ClientType:     payment.ClientType(req.ClientType),
		ClientIp:       req.ClientIp,
		BizOrderNo:     req.BizOrderNo,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateCryptoRechargeOrderResp{
		RespBase: respBase(result.Base),
		Data:     rechargeOrderFromPB(result.Order),
		Address:  cryptoRechargeAddressFromPB(result.Address),
	}, nil
}
