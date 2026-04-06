package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineLogic {
	return &GetKlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取K线
func (l *GetKlineLogic) GetKline(in *itick.GetKlineReq) (*itick.GetKlineResp, error) {
	model := l.svcCtx.Factory.New(in.CategoryCode, l.buildInterval(in.KType))
	if model == nil {
		return &itick.GetKlineResp{}, nil
	}
	result, err := model.FindBeforeTsBySymbol(l.ctx, in.Symbol, in.EndTs, in.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.Kline, 0)
	for _, item := range result {
		data = append(data, &itick.Kline{
			CategoryCode: item.CategoryCode,
			Market:       item.Market,
			Symbol:       item.Symbol,
			KType:        in.KType,
			Ts:           item.Ts,
			Open:         item.Open,
			High:         item.High,
			Low:          item.Low,
			Close:        item.Close,
			Volume:       item.Volume,
			Turnover:     item.Turnover,
		})
	}
	return &itick.GetKlineResp{
		Base: &itick.RespBase{
			Code: 200,
			Msg:  "获取kline成功",
		},
		Data: data,
	}, nil
}

func (l *GetKlineLogic) buildInterval(kType itick.KlineType) string {
	switch kType {
	case itick.KlineType_KLINE_TYPE_1M:
		return "1m"
	case itick.KlineType_KLINE_TYPE_5M:
		return "5m"
	case itick.KlineType_KLINE_TYPE_15M:
		return "15m"
	case itick.KlineType_KLINE_TYPE_30M:
		return "30m"
	case itick.KlineType_KLINE_TYPE_60M:
		return "1h"
	case itick.KlineType_KLINE_TYPE_1D:
		return "1d"
	case itick.KlineType_KLINE_TYPE_1W:
		return "1w"
	case itick.KlineType_KLINE_TYPE_1MO:
		return "1mo"
	default:
		return "1m"
	}
}
