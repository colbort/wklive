package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyCryptoRechargeTxsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyCryptoRechargeTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCryptoRechargeTxsLogic {
	return &ListMyCryptoRechargeTxsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 我的链上充值交易列表
func (l *ListMyCryptoRechargeTxsLogic) ListMyCryptoRechargeTxs(in *payment.ListMyCryptoRechargeTxsReq) (*payment.ListMyCryptoRechargeTxsResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := listCryptoRechargeTxs(l.ctx, l.svcCtx, listCryptoTxReq{
		tenantId:        tenantId,
		userId:          userId,
		orderNo:         in.OrderNo,
		coin:            in.Coin,
		chainCode:       in.ChainCode,
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
	return &payment.ListMyCryptoRechargeTxsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastCryptoTxID(items)),
		Data: data,
	}, nil
}
