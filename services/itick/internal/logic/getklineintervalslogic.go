package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlineIntervalsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetKlineIntervalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineIntervalsLogic {
	return &GetKlineIntervalsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 kline 粒度
func (l *GetKlineIntervalsLogic) GetKlineIntervals(in *itick.AppEmpty) (*itick.KlineIntervalsResp, error) {
	return &itick.KlineIntervalsResp{
		Data: []*itick.KlineInterval{
			{Name: "1m", KType: 1},
			{Name: "5m", KType: 2},
			{Name: "15m", KType: 3},
			{Name: "30m", KType: 4},
			{Name: "1h", KType: 5},
			{Name: "1d", KType: 8},
			{Name: "1w", KType: 9},
			{Name: "1mo", KType: 10},
		},
	}, nil
}
