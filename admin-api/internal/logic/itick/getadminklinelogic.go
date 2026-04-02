// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminKlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAdminKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminKlineLogic {
	return &GetAdminKlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAdminKlineLogic) GetAdminKline(req *types.GetAdminKlineReq) (resp *types.GetAdminKlineResp, err error) {
	result, err := l.svcCtx.ItickCli.GetAdminKline(l.ctx, &itick.GetAdminKlineReq{
		Market: req.Market,
		Symbol: req.Symbol,
		KType:  itick.KlineType(req.KType),
		EndTs:  req.EndTs,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.Kline, 0)
	for _, item := range result.Data {
		data = append(data, types.Kline{
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
	return &types.GetAdminKlineResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
