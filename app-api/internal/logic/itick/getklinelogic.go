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

type GetKlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineLogic {
	return &GetKlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKlineLogic) GetKline(req *types.GetKlineReq) (resp *types.GetKlineResp, err error) {
	result, err := l.svcCtx.ItickCli.GetKline(l.ctx, &itick.GetKlineReq{
		Market: req.Market,
		Symbol: req.Symbol,
		KType:  itick.KlineType(req.KType),
		EndTs:  req.EndTs,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetKlineResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: make([]types.Kline, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		resp.Data = append(resp.Data, types.Kline{
			Market:   item.Market,
			Symbol:   item.Symbol,
			KType:    int64(item.KType),
			Ts:       item.Ts,
			Open:     item.Open,
			High:     item.High,
			Low:      item.Low,
			Close:    item.Close,
			Volume:   item.Volume,
			Turnover: item.Turnover,
		})
	}

	return
}
