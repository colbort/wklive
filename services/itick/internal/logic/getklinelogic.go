package logic

import (
	"context"
	"time"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"

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
