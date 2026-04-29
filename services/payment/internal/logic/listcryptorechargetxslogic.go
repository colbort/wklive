package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCryptoRechargeTxsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCryptoRechargeTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCryptoRechargeTxsLogic {
	return &ListCryptoRechargeTxsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 链上充值交易列表
func (l *ListCryptoRechargeTxsLogic) ListCryptoRechargeTxs(in *payment.ListCryptoRechargeTxsReq) (*payment.ListCryptoRechargeTxsResp, error) {
	items, total, err := listCryptoRechargeTxs(l.ctx, l.svcCtx, listCryptoTxReq{
		tenantId:        in.TenantId,
		userId:          in.UserId,
		orderNo:         in.OrderNo,
		coin:            in.Coin,
		chainCode:       in.ChainCode,
		txHash:          in.TxHash,
		toAddress:       in.ToAddress,
		status:          in.Status,
		createTimeStart: in.CreateTimeStart,
		createTimeEnd:   in.CreateTimeEnd,
		cursor:          in.Page.Cursor,
		limit:           in.Page.Limit,
	})
	if err != nil {
		return nil, err
	}
	data := make([]*payment.CryptoRechargeTx, 0, len(items))
	for _, item := range items {
		data = append(data, toCryptoRechargeTxProto(item))
	}
	return &payment.ListCryptoRechargeTxsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastCryptoTxID(items)),
		Data: data,
	}, nil
}
