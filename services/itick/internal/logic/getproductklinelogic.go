package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

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
func (l *GetProductKlineLogic) GetProductKline(in *itick.GetProductKlineReq) (*itick.GetProductKlineResp, error) {
	getKlineLogic := NewGetKlineLogic(l.ctx, l.svcCtx)
	result, err := getKlineLogic.GetKline(&itick.GetKlineReq{
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

	return &itick.GetProductKlineResp{
		Base: result.Base,
		Data: result.Data,
	}, nil
}
