package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailLogic {
	return &GetSymbolDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取指定交易对详情
func (l *GetSymbolDetailLogic) GetSymbolDetail(in *trade.GetSymbolDetailReq) (*trade.GetSymbolDetailResp, error) {
	item, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetSymbolDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := &trade.GetSymbolDetailResp{
		Base:   helper.OkResp(),
		Symbol: symbolToProto(item),
	}
	spot, err := l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, in.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if spot != nil {
		resp.Spot = spotSymbolToProto(spot)
	}
	contractCfg, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, in.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if contractCfg != nil {
		resp.Contract = contractSymbolToProto(contractCfg)
	}
	return resp, nil
}
