package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeOrderModel.FindPage(
		l.ctx,
		models.StakeOrderPageFilter{
			TenantId:    in.TenantId,
			UserId:      in.UserId,
			ProductId:   in.ProductId,
			OrderNo:     in.OrderNo,
			ProductName: in.ProductName,
			CoinSymbol:  in.CoinSymbol,
			Status:      int64(in.Status),
			RedeemType:  int64(in.RedeemType),
			Source:      int64(in.Source),
			StartBegin:  in.StartTimesBegin,
			StartEnd:    in.StartTimesEnd,
			EndBegin:    in.EndTimesBegin,
			EndEnd:      in.EndTimesEnd,
		},
		cursor,
		limit,
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
