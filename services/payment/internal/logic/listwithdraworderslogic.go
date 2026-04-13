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

type ListWithdrawOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawOrdersLogic {
	return &ListWithdrawOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现订单列表
func (l *ListWithdrawOrdersLogic) ListWithdrawOrders(in *payment.ListWithdrawOrdersReq) (*payment.ListWithdrawOrdersResp, error) {
	orders, total, err := l.svcCtx.WithdrawOrderModel.FindPage(
		l.ctx,
		in.TenantId,
		in.UserId,
		in.OrderNo,
		0,
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

	data := make([]*payment.WithdrawOrder, 0, len(orders))
	for _, o := range orders {
		data = append(data, toWithdrawOrderProto(o))
	}

	return &payment.ListWithdrawOrdersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(orders), total, lastID),
		Data: data,
	}, nil
}
