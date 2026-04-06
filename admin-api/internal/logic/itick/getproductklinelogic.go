// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductKlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductKlineLogic {
	return &GetProductKlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductKlineLogic) GetProductKline(req *types.GetProductKlineReq) (resp *types.GetProductKlineResp, err error) {
	result, err := l.svcCtx.ItickCli.GetProductKline(l.ctx, &itick.GetProductKlineReq{
		CategoryCode: req.CategoryCode,
		Market:       req.Market,
		Symbol:       req.Symbol,
		KType:        itick.KlineType(req.KType),
		EndTs:        req.EndTs,
		Limit:        req.Limit,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.Kline, 0)
	for _, item := range result.Data {
		data = append(data, types.Kline{
			CategoryCode: item.CategoryCode,
			Market:       item.Market,
			Symbol:       item.Symbol,
			KType:        int64(item.KType),
			Ts:           item.Ts,
			Open:         item.Open,
			High:         item.High,
			Low:          item.Low,
			Close:        item.Close,
			Volume:       item.Volume,
			Turnover:     item.Turnover,
		})
	}
	return &types.GetProductKlineResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
