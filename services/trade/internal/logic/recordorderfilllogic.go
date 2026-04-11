package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordOrderFillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordOrderFillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordOrderFillLogic {
	return &RecordOrderFillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 记录订单成交信息
func (l *RecordOrderFillLogic) RecordOrderFill(in *trade.RecordOrderFillReq) (*trade.InternalCommonResp, error) {
	if in.Fill == nil {
		return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
	}
	exists, err := l.svcCtx.TradeFillModel.FindOneByTenantIdFillNo(l.ctx, in.Fill.TenantId, in.Fill.FillNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exists == nil {
		_, err = l.svcCtx.TradeFillModel.Insert(l.ctx, &models.TTradeFill{
			TenantId:      in.Fill.TenantId,
			FillNo:        in.Fill.FillNo,
			OrderId:       in.Fill.OrderId,
			OrderNo:       in.Fill.OrderNo,
			UserId:        in.Fill.UserId,
			SymbolId:      in.Fill.SymbolId,
			MarketType:    int64(in.Fill.MarketType),
			Side:          int64(in.Fill.Side),
			PositionSide:  int64(in.Fill.PositionSide),
			Price:         mustParseFloat(in.Fill.Price),
			Qty:           mustParseFloat(in.Fill.Qty),
			Amount:        mustParseFloat(in.Fill.Amount),
			Fee:           mustParseFloat(in.Fill.Fee),
			FeeAsset:      in.Fill.FeeAsset,
			LiquidityType: int64(in.Fill.LiquidityType),
			RealizedPnl:   mustParseFloat(in.Fill.RealizedPnl),
			MatchTime:     in.Fill.MatchTime,
			CreateTimes:   in.Fill.CreateTimes,
		})
		if err != nil {
			return nil, err
		}
	}
	return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
}
