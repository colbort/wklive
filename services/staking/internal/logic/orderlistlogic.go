package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押订单列表
func (l *OrderListLogic) OrderList(in *staking.AdminOrderListReq) (*staking.AdminOrderListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeOrderModel.FindPage(
		l.ctx, in.TenantId, cursor, limit, in.Uid, in.ProductId, in.OrderNo, in.ProductName, in.CoinSymbol,
		int64(in.Status), int64(in.RedeemType), int64(in.Source),
		in.StartTimesBegin, in.StartTimesEnd, in.EndTimesBegin, in.EndTimesEnd,
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.AdminOrderListResp{Page: helper.OkResp()}
	if len(items) == 0 {
		resp.Page = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeOrder, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, orderToProto(item))
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
