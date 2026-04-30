package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyCryptoRechargeTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeTxLogic {
	return &GetMyCryptoRechargeTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyCryptoRechargeTxLogic) GetMyCryptoRechargeTx(req *types.GetMyCryptoRechargeTxReq) (*types.GetMyCryptoRechargeTxResp, error) {
	result, err := l.svcCtx.PaymentCli.GetMyCryptoRechargeTx(l.ctx, &payment.GetMyCryptoRechargeTxReq{
		Id:     req.Id,
		TxHash: req.TxHash,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetMyCryptoRechargeTxResp{
		RespBase: respBase(result.Base),
		Data:     cryptoRechargeTxFromPB(result.Data),
	}, nil
}
