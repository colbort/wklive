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

type MyOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyOrderListLogic {
	return &MyOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的质押订单列表
func (l *MyOrderListLogic) MyOrderList(in *staking.AppMyOrderListReq) (*staking.AppMyOrderListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.StakeOrderModel.FindPage(l.ctx, models.StakeOrderPageFilter{
		TenantId:   tenantId,
		UserId:     userId,
		Status:     int64(in.Status),
		RedeemType: int64(in.RedeemType),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}

	resp := &staking.AppMyOrderListResp{Base: helper.OkResp()}
	if len(items) == 0 {
		resp.Base = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeOrder, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, orderToProto(item))
	}
	resp.Base = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
