package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyRechargeOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyRechargeOrdersLogic {
	return &ListMyRechargeOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的充值订单列表
func (l *ListMyRechargeOrdersLogic) ListMyRechargeOrders(in *payment.ListMyRechargeOrdersReq) (*payment.ListMyRechargeOrdersResp, error) {
	items, total, err := l.svcCtx.RechargeOrderModel.FindPage(l.ctx, in.TenantId, in.UserId, in.OrderNo, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*payment.RechargeOrder, 0)
	for _, order := range items {
		data = append(data, toRechargeOrderProto(order))
	}

	return &payment.ListMyRechargeOrdersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		List: data,
	}, nil
}
