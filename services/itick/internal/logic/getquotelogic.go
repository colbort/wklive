package logic

import (
	"context"
	"fmt"

	"wklive/proto/itick"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteLogic {
	return &GetQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取最新报价
func (l *GetQuoteLogic) GetQuote(in *itick.GetQuoteReq) (*itick.GetQuoteResp, error) {
	key := fmt.Sprintf("itick:quote:%s:%s:%s", in.CategoryCode, in.Market, in.Symbol)
	var data client.QuotePayload
	err := l.svcCtx.Cache.GetCtx(l.ctx, key, &data)
	if err != nil {
		return nil, err
	}

	return &itick.GetQuoteResp{
		Base: &itick.RespBase{
			Code: 200,
			Msg:  "获取报价成功",
		},
		Data: &itick.Quote{
			CategoryCode: in.CategoryCode,
			Market:       in.Market,
			Symbol:       in.Symbol,
			LastPrice:    data.LastPrice,
			OpenPrice:    data.Open,
			HighPrice:    data.High,
			LowPrice:     data.LastPrice,
			Volume:       data.Volume,
			Turnover:     data.Turnover,
			QuoteTs:      data.Ts,
		},
	}, nil
}
