package applogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/proto/market"
	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetQuoteLogic {
	return &BatchGetQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量获取最新报价
func (l *BatchGetQuoteLogic) BatchGetQuote(in *market.BatchGetQuoteReq) (*market.BatchGetQuoteResp, error) {
	msgs := make([]types.ClientMessage, 0, len(in.GetData()))
	for _, item := range in.GetData() {
		if item == nil {
			continue
		}
		categoryCode := item.GetCategoryCode()
		if categoryCode == "" {
			categoryCode = in.GetCategoryCode()
		}
		market := item.GetMarket()
		if market == "" {
			market = in.GetMarket()
		}
		msg := cache.NormalizeClientMessage(types.ClientMessage{
			Topic:        types.TopicQuote,
			CategoryCode: categoryCode,
			Market:       market,
			Symbol:       item.GetSymbol(),
		})
		if msg.CategoryCode == "" || msg.Market == "" || msg.Symbol == "" {
			continue
		}
		msgs = append(msgs, msg)
	}

	result, err := l.svcCtx.MarketDataCache.ReadMany(l.ctx, msgs)
	if err != nil {
		return nil, err
	}
	data := make([]*market.Quote, 0, len(result))
	for _, item := range result {
		quote, ok := item.Payload.(*types.QuotePayload)
		if !ok || quote == nil {
			continue
		}
		data = append(data, helpers.ToQuotePayloadProto(item.Message.CategoryCode, item.Message.Market, item.Message.Symbol, quote))
	}

	return &market.BatchGetQuoteResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
