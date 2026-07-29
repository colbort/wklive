package adminlogic

import (
	"context"

	"wklive/proto/market"
	marketapplogic "wklive/services/market/internal/logic/app"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductKlineLogic {
	return &GetProductKlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// K线查看
func (l *GetProductKlineLogic) GetProductKline(in *market.GetProductKlineReq) (*market.GetProductKlineResp, error) {
	getKlineLogic := marketapplogic.NewGetKlineLogic(l.ctx, l.svcCtx)
	result, err := getKlineLogic.GetKline(&market.GetKlineReq{
		CategoryCode: in.CategoryCode,
		Market:       in.Market,
		Symbol:       in.Symbol,
		KType:        in.KType,
		EndTs:        in.EndTs,
		Limit:        in.Limit,
	})
	if err != nil {
		return nil, err
	}

	return &market.GetProductKlineResp{
		Base: result.Base,
		Data: result.Data,
	}, nil
}
