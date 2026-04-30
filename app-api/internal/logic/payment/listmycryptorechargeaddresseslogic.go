package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyCryptoRechargeAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeAddressesLogic {
	return &ListMyCryptoRechargeAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyCryptoRechargeAddressesLogic) ListMyCryptoRechargeAddresses(req *types.ListMyCryptoRechargeAddressesReq) (*types.ListMyCryptoRechargeAddressesResp, error) {
	result, err := l.svcCtx.PaymentCli.ListMyCryptoRechargeAddresses(l.ctx, &payment.ListMyCryptoRechargeAddressesReq{
		WalletType: req.WalletType,
		Coin:       req.Coin,
		ChainCode:  common.ChainCode(req.ChainCode),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.CryptoRechargeAddress, 0, len(result.Data))
	for _, item := range result.Data {
		data = append(data, cryptoRechargeAddressFromPB(item))
	}

	return &types.ListMyCryptoRechargeAddressesResp{
		RespBase: respBase(result.Base),
		Data:     data,
	}, nil
}
