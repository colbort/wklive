package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordPositionHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordPositionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordPositionHistoryLogic {
	return &RecordPositionHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 记录持仓历史信息
func (l *RecordPositionHistoryLogic) RecordPositionHistory(in *trade.RecordPositionHistoryReq) (*trade.InternalCommonResp, error) {
	if in.History == nil {
		return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
	}
	_, err := l.svcCtx.ContractPositionHistModel.Insert(l.ctx, &models.TContractPositionHistory{
		TenantId:             in.History.TenantId,
		PositionId:           in.History.PositionId,
		UserId:               in.History.UserId,
		SymbolId:             in.History.SymbolId,
		MarketType:           int64(in.History.MarketType),
		PositionSide:         int64(in.History.PositionSide),
		ActionType:           int64(in.History.ActionType),
		BeforeQty:            mustParseFloat(in.History.BeforeQty),
		AfterQty:             mustParseFloat(in.History.AfterQty),
		BeforeAvailQty:       mustParseFloat(in.History.BeforeAvailQty),
		AfterAvailQty:        mustParseFloat(in.History.AfterAvailQty),
		BeforeFrozenQty:      mustParseFloat(in.History.BeforeFrozenQty),
		AfterFrozenQty:       mustParseFloat(in.History.AfterFrozenQty),
		BeforeOpenAvgPrice:   mustParseFloat(in.History.BeforeOpenAvgPrice),
		AfterOpenAvgPrice:    mustParseFloat(in.History.AfterOpenAvgPrice),
		BeforePositionMargin: mustParseFloat(in.History.BeforePositionMargin),
		AfterPositionMargin:  mustParseFloat(in.History.AfterPositionMargin),
		BeforeIsolatedMargin: mustParseFloat(in.History.BeforeIsolatedMargin),
		AfterIsolatedMargin:  mustParseFloat(in.History.AfterIsolatedMargin),
		BeforeUnrealizedPnl:  mustParseFloat(in.History.BeforeUnrealizedPnl),
		AfterUnrealizedPnl:   mustParseFloat(in.History.AfterUnrealizedPnl),
		RealizedPnlDelta:     mustParseFloat(in.History.RealizedPnlDelta),
		FeeDelta:             mustParseFloat(in.History.FeeDelta),
		FeeAsset:             in.History.FeeAsset,
		MarkPrice:            mustParseFloat(in.History.MarkPrice),
		RefOrderId:           in.History.RefOrderId,
		RefFillId:            in.History.RefFillId,
		OperatorId:           in.History.OperatorId,
		Source:               int64(in.History.Source),
		Remark:               in.History.Remark,
		CreateTimes:          in.History.CreateTimes,
	})
	if err != nil {
		return nil, err
	}
	return &trade.InternalCommonResp{Base: helper.OkResp()}, nil
}
