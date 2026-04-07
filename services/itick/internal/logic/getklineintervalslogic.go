package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
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

	data := make([]*itick.KlineInterval, 0, len(utils.KlineIntervals))
	for _, v := range utils.KlineIntervals {
		data = append(data, &itick.KlineInterval{
			Name:  v.Name,
			KType: int32(v.KType), // 看你 proto 这里类型是不是 int32
		})
	}
	return &itick.KlineIntervalsResp{
		Data: data,
	}, nil
}
