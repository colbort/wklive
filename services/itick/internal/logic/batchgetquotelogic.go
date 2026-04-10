package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

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
func (l *BatchGetQuoteLogic) BatchGetQuote(in *itick.BatchGetQuoteReq) (*itick.BatchGetQuoteResp, error) {
	result, err := l.svcCtx.ItickQuoteModel.FindQuotes(l.ctx, in.Data)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.Quote, 0)
	for _, quote := range result {
		data = append(data, &itick.Quote{
			CategoryCode:   quote.CategoryCode,
			Market:         quote.Market,
			Symbol:         quote.Symbol,
			LastPrice:      quote.LastPrice,
			OpenPrice:      quote.OpenPrice,
			HighPrice:      quote.HighPrice,
			LowPrice:       quote.LowPrice,
			PrevClosePrice: quote.PrevClosePrice,
			ChangeValue:    quote.ChangeValue,
			ChangeRate:     quote.ChangeRate,
			Volume:         quote.Volume,
			Turnover:       quote.Turnover,
			QuoteTs:        quote.QuoteTs,
			TradeStatus:    quote.TradeStatus,
		})
	}

	return &itick.BatchGetQuoteResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
