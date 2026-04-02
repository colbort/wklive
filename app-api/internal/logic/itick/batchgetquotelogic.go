// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetQuoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetQuoteLogic {
	return &BatchGetQuoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetQuoteLogic) BatchGetQuote(req *types.BatchGetQuoteReq) (resp *types.BatchGetQuoteResp, err error) {
	data := make([]*itick.MarketSymbol, 0, len(req.Data))
	for _, item := range req.Data {
		data = append(data, &itick.MarketSymbol{
			Market: item.Market,
			Symbol: item.Symbol,
		})
	}

	result, err := l.svcCtx.ItickCli.BatchGetQuote(l.ctx, &itick.BatchGetQuoteReq{
		Market: req.Market,
		Data:   data,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.BatchGetQuoteResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: make([]types.Quote, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		resp.Data = append(resp.Data, types.Quote{
			Market:         item.Market,
			Symbol:         item.Symbol,
			LastPrice:      item.LastPrice,
			OpenPrice:      item.OpenPrice,
			HighPrice:      item.HighPrice,
			LowPrice:       item.LowPrice,
			PrevClosePrice: item.PrevClosePrice,
			ChangeValue:    item.ChangeValue,
			ChangeRate:     item.ChangeRate,
			Volume:         item.Volume,
			Turnover:       item.Turnover,
			QuoteTs:        item.QuoteTs,
			TradeStatus:    int64(item.TradeStatus),
		})
	}

	return
}
