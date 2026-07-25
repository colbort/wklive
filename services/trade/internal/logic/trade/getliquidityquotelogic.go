package tradelogic

import (
	"context"
	"fmt"

	"wklive/proto/trade"
	applogic "wklive/services/trade/internal/logic/app"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLiquidityQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLiquidityQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLiquidityQuoteLogic {
	return &GetLiquidityQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLiquidityQuoteLogic) GetLiquidityQuote(in *trade.GetLiquidityQuoteReq) (*trade.GetOrderDetailResp, error) {
	if in.GetOrder() == nil || in.TradeUserId <= 0 || in.Order.OrderId <= 0 {
		return nil, fmt.Errorf("trade_user_id and order_id are required")
	}
	order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, in.Order.OrderId)
	if err != nil {
		return nil, err
	}
	ctx, err := liquidityUserContext(l.ctx, order.TenantId, in.TradeUserId)
	if err != nil {
		return nil, err
	}
	return applogic.NewGetOrderDetailLogic(ctx, l.svcCtx).GetOrderDetail(in.Order)
}
