package tradelogic

import (
	"context"
	"fmt"

	"wklive/proto/trade"
	applogic "wklive/services/trade/internal/logic/app"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlaceLiquidityQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceLiquidityQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceLiquidityQuoteLogic {
	return &PlaceLiquidityQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 平台内部做市账户报价，不依赖用户登录 metadata。
func (l *PlaceLiquidityQuoteLogic) PlaceLiquidityQuote(in *trade.PlaceLiquidityQuoteReq) (*trade.PlaceOrderResp, error) {
	if in.GetOrder() == nil || in.TradeUserId <= 0 {
		return nil, fmt.Errorf("trade_user_id and order are required")
	}
	tenantID, err := symbolTenant(l.ctx, l.svcCtx, in.Order.SymbolId)
	if err != nil {
		return nil, err
	}
	ctx, err := liquidityUserContext(l.ctx, tenantID, in.TradeUserId)
	if err != nil {
		return nil, err
	}
	return applogic.NewPlaceOrderLogic(ctx, l.svcCtx).PlaceOrder(in.Order)
}
