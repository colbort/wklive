package internallogic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func writeSystemPositionHistory(ctx context.Context, model models.TContractPositionHistoryModel, before, after *models.TContractPosition, actionKey string, action trade.PositionActionType, realized, fee, mark decimal.Decimal, remark string) error {
	_, err := model.Insert(ctx, &models.TContractPositionHistory{TenantId: after.TenantId, PositionId: after.Id, UserId: after.UserId, SymbolId: after.SymbolId, ContractType: after.ContractType, ContractValueType: after.ContractValueType, PositionSide: after.PositionSide, ActionType: int64(action), ActionKey: actionKey, BeforeQty: before.Qty, AfterQty: after.Qty, BeforeAvailQty: before.AvailQty, AfterAvailQty: after.AvailQty, BeforeFrozenQty: before.FrozenQty, AfterFrozenQty: after.FrozenQty, BeforeOpenAvgPrice: before.OpenAvgPrice, AfterOpenAvgPrice: after.OpenAvgPrice, BeforePositionMargin: before.PositionMargin, AfterPositionMargin: after.PositionMargin, BeforeIsolatedMargin: before.IsolatedMargin, AfterIsolatedMargin: after.IsolatedMargin, BeforeUnrealizedPnl: before.UnrealizedPnl, AfterUnrealizedPnl: after.UnrealizedPnl, RealizedPnlDelta: realized, FeeDelta: fee, FeeAsset: after.MarginAsset, MarkPrice: mark, Source: int64(trade.SourceType_SOURCE_TYPE_TASK), Remark: remark, CreateTimes: utils.NowMillis()})
	return err
}
