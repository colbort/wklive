package tradelogic

import (
	"context"
	"fmt"

	"wklive/proto/trade"
	applogic "wklive/services/trade/internal/logic/app"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelLiquidityQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelLiquidityQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLiquidityQuoteLogic {
	return &CancelLiquidityQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelLiquidityQuoteLogic) CancelLiquidityQuote(in *trade.CancelLiquidityQuoteReq) (*trade.UserCommonResp, error) {
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
	return applogic.NewCancelOrderLogic(ctx, l.svcCtx).CancelOrder(in.Order)
}
