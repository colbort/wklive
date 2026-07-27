package tradelogic

import (
	"context"
	"fmt"
	"wklive/services/trade/internal/logic/helpers"

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
	actionKey := fmt.Sprintf("manual:%d:%d:%d:%d", in.History.PositionId, in.History.ActionType, in.History.OperatorId, in.History.CreateTimes)
	if in.History.RefFillId > 0 {
		actionKey = fmt.Sprintf("fill:%d:%d:%d", in.History.RefFillId, in.History.PositionId, in.History.ActionType)
	} else if in.History.RefOrderId > 0 {
		actionKey = fmt.Sprintf("order:%d:%d:%d", in.History.RefOrderId, in.History.PositionId, in.History.ActionType)
	}
	_, err := l.svcCtx.ContractPositionHistModel.Insert(l.ctx, &models.TContractPositionHistory{
		TenantId:             in.History.TenantId,
		PositionId:           in.History.PositionId,
		UserId:               in.History.UserId,
		SymbolId:             in.History.SymbolId,
		ContractType:         int64(in.History.ContractType),
		ContractValueType:    int64(in.History.ContractValueType),
		PositionSide:         int64(in.History.PositionSide),
		ActionType:           int64(in.History.ActionType),
		ActionKey:            actionKey,
		BeforeQty:            helpers.MustParseFloat(in.History.BeforeQty),
		AfterQty:             helpers.MustParseFloat(in.History.AfterQty),
		BeforeAvailQty:       helpers.MustParseFloat(in.History.BeforeAvailQty),
		AfterAvailQty:        helpers.MustParseFloat(in.History.AfterAvailQty),
		BeforeFrozenQty:      helpers.MustParseFloat(in.History.BeforeFrozenQty),
		AfterFrozenQty:       helpers.MustParseFloat(in.History.AfterFrozenQty),
		BeforeOpenAvgPrice:   helpers.MustParseFloat(in.History.BeforeOpenAvgPrice),
		AfterOpenAvgPrice:    helpers.MustParseFloat(in.History.AfterOpenAvgPrice),
		BeforePositionMargin: helpers.MustParseFloat(in.History.BeforePositionMargin),
		AfterPositionMargin:  helpers.MustParseFloat(in.History.AfterPositionMargin),
		BeforeIsolatedMargin: helpers.MustParseFloat(in.History.BeforeIsolatedMargin),
		AfterIsolatedMargin:  helpers.MustParseFloat(in.History.AfterIsolatedMargin),
		BeforeUnrealizedPnl:  helpers.MustParseFloat(in.History.BeforeUnrealizedPnl),
		AfterUnrealizedPnl:   helpers.MustParseFloat(in.History.AfterUnrealizedPnl),
		RealizedPnlDelta:     helpers.MustParseFloat(in.History.RealizedPnlDelta),
		FeeDelta:             helpers.MustParseFloat(in.History.FeeDelta),
		FeeAsset:             in.History.FeeAsset,
		MarkPrice:            helpers.MustParseFloat(in.History.MarkPrice),
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
