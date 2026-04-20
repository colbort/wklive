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

type GetQuoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteLogic {
	return &GetQuoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQuoteLogic) GetQuote(req *types.GetQuoteReq) (resp *types.GetQuoteResp, err error) {
	result, err := l.svcCtx.ItickCli.GetQuote(l.ctx, &itick.GetQuoteReq{
		CategoryCode: req.CategoryCode,
		Market:       req.Market,
		Symbol:       req.Symbol,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetQuoteResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.Quote{
			CategoryCode:   result.Data.CategoryCode,
			Market:         result.Data.Market,
			Symbol:         result.Data.Symbol,
			LastPrice:      result.Data.LastPrice,
			OpenPrice:      result.Data.OpenPrice,
			HighPrice:      result.Data.HighPrice,
			LowPrice:       result.Data.LowPrice,
			PrevClosePrice: result.Data.PrevClosePrice,
			ChangeValue:    result.Data.ChangeValue,
			ChangeRate:     result.Data.ChangeRate,
			Volume:         result.Data.Volume,
			Turnover:       result.Data.Turnover,
			QuoteTs:        result.Data.QuoteTs,
			TradeStatus:    int64(result.Data.TradeStatus),
		},
	}

	return
}
