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

type ListRechargeOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRechargeOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeOrdersLogic {
	return &ListRechargeOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值订单列表
func (l *ListRechargeOrdersLogic) ListRechargeOrders(in *payment.ListRechargeOrdersReq) (*payment.ListRechargeOrdersResp, error) {
	orders, total, err := l.svcCtx.RechargeOrderModel.FindPage(
		l.ctx,
		in.TenantId,
		in.UserId,
		in.OrderNo,
		int64(in.Status),
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(orders) > 0 {
		lastID = orders[len(orders)-1].Id
	}

	data := make([]*payment.RechargeOrder, 0, len(orders))
	for _, o := range orders {
		data = append(data, toRechargeOrderProto(o))
	}

	return &payment.ListRechargeOrdersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(orders), total, lastID),
		Data: data,
	}, nil
}
