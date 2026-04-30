package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyCryptoRechargeTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeTxsLogic {
	return &ListMyCryptoRechargeTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyCryptoRechargeTxsLogic) ListMyCryptoRechargeTxs(req *types.ListMyCryptoRechargeTxsReq) (*types.ListMyCryptoRechargeTxsResp, error) {
	result, err := l.svcCtx.PaymentCli.ListMyCryptoRechargeTxs(l.ctx, &payment.ListMyCryptoRechargeTxsReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		OrderNo:         req.OrderNo,
		Coin:            req.Coin,
		ChainCode:       common.ChainCode(req.ChainCode),
		Status:          payment.CryptoRechargeTxStatus(req.Status),
		CreateTimeStart: req.CreateTimeStart,
		CreateTimeEnd:   req.CreateTimeEnd,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.CryptoRechargeTx, 0, len(result.Data))
	for _, item := range result.Data {
		data = append(data, cryptoRechargeTxFromPB(item))
	}

	return &types.ListMyCryptoRechargeTxsResp{
		RespBase: respBase(result.Base),
		Data:     data,
	}, nil
}
