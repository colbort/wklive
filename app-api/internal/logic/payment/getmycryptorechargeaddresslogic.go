package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeAddressLogic {
	return &GetMyCryptoRechargeAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyCryptoRechargeAddressLogic) GetMyCryptoRechargeAddress(req *types.GetMyCryptoRechargeAddressReq) (*types.GetMyCryptoRechargeAddressResp, error) {
	result, err := l.svcCtx.PaymentCli.GetMyCryptoRechargeAddress(l.ctx, &payment.GetMyCryptoRechargeAddressReq{
		WalletType: req.WalletType,
		Coin:       req.Coin,
		ChainCode:  common.ChainCode(req.ChainCode),
	})
	if err != nil {
		return nil, err
	}

	return &types.GetMyCryptoRechargeAddressResp{
		RespBase: respBase(result.Base),
		Data:     cryptoRechargeAddressFromPB(result.Data),
	}, nil
}
