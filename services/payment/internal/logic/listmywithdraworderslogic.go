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

type ListMyWithdrawOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyWithdrawOrdersLogic {
	return &ListMyWithdrawOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的提现订单列表
func (l *ListMyWithdrawOrdersLogic) ListMyWithdrawOrders(in *payment.ListMyWithdrawOrdersReq) (*payment.ListMyWithdrawOrdersResp, error) {
	items, total, err := l.svcCtx.WithdrawOrderModel.FindPage(l.ctx, in.TenantId, in.UserId, "", 0, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*payment.WithdrawOrder, 0)
	for _, order := range items {
		data = append(data, toWithdrawOrderProto(order))
	}

	return &payment.ListMyWithdrawOrdersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
