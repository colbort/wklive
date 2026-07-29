package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPriceFormulaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPriceFormulaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceFormulaLogic {
	return &GetPriceFormulaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPriceFormulaLogic) GetPriceFormula(in *market.PriceFormulaReq) (*market.PriceFormulaResp, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("price formula id is required")
	}
	row, err := l.svcCtx.PriceFormulaModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &market.PriceFormulaResp{Base: helper.OkResp(), Data: toPriceFormulaProto(row)}, nil
}
