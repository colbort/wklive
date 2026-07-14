package logic

import (
	"context"
	"sort"
	"time"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineLogic {
	return &GetKlineLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetKline is an App read path. It only reads MongoDB and never calls iTick REST.
func (l *GetKlineLogic) GetKline(in *itick.GetKlineReq) (*itick.GetKlineResp, error) {
	interval := utils.KlineTypeToInterval(in.KType)
	if interval == "1y" {
		return l.getYearKlines(in)
	}
	model := l.svcCtx.Factory.New(in.CategoryCode, interval)
	if model == nil {
		return &itick.GetKlineResp{Base: helper.OkResp(), Data: []*itick.Kline{}}, nil
	}
	endTs := in.EndTs
	if endTs <= 0 {
		endTs = time.Now().UnixMilli() + 1
	}
	result, err := model.FindBeforeTsByMarketSymbol(l.ctx, in.Market, in.Symbol, endTs, in.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.Kline, 0, len(result))
	for _, item := range result {
		data = append(data, toKlineProto(in.KType, item))
	}
	return &itick.GetKlineResp{Base: helper.OkResp(), Data: data}, nil
}

func (l *GetKlineLogic) getYearKlines(in *itick.GetKlineReq) (*itick.GetKlineResp, error) {
	endTs := in.EndTs
	if endTs <= 0 {
		endTs = time.Now().UnixMilli() + 1
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 100
	}
	daily := l.svcCtx.Factory.New(in.CategoryCode, "1d")
	if daily == nil {
		return &itick.GetKlineResp{Base: helper.OkResp(), Data: []*itick.Kline{}}, nil
	}
	rows, err := daily.FindBeforeTsByMarketSymbol(l.ctx, in.Market, in.Symbol, endTs, limit*370)
	if err != nil {
		return nil, err
	}
	byYear := make(map[int][]*models.CoinKline)
	for _, row := range rows {
		year := time.UnixMilli(row.Ts).UTC().Year()
		byYear[year] = append(byYear[year], row)
	}
	years := make([]int, 0, len(byYear))
	for year := range byYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	data := make([]*itick.Kline, 0, min(int(limit), len(years)))
	for _, year := range years {
		if int64(len(data)) >= limit {
			break
		}
		list := byYear[year]
		sort.Slice(list, func(i, j int) bool { return list[i].Ts < list[j].Ts })
		bar := aggregateQueryKlines(in, time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(), list)
		data = append(data, toKlineProto(in.KType, bar))
	}
	return &itick.GetKlineResp{Base: helper.OkResp(), Data: data}, nil
}

func aggregateQueryKlines(in *itick.GetKlineReq, ts int64, list []*models.CoinKline) *models.CoinKline {
	bar := &models.CoinKline{CategoryCode: in.CategoryCode, Market: in.Market, Symbol: in.Symbol,
		Interval: "1y", Ts: ts, Open: list[0].Open, High: list[0].High, Low: list[0].Low,
		Close: list[len(list)-1].Close, Source: models.KlineSourceDerived,
		IsClosed: time.Now().UTC().Year() > time.UnixMilli(ts).UTC().Year(), ActualCount: int32(len(list))}
	bar.Confirmed = bar.IsClosed
	for _, item := range list {
		bar.High = max(bar.High, item.High)
		bar.Low = min(bar.Low, item.Low)
		bar.Volume += item.Volume
		bar.Turnover += item.Turnover
		bar.Confirmed = bar.Confirmed && item.Confirmed
	}
	return bar
}
