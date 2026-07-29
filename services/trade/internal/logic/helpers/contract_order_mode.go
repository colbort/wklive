package helpers

import (
	"context"
	"errors"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

// ValidateContractOpeningMode prevents an exposure-increasing order from
// bypassing a configured position/margin mode or mixing incompatible active
// risk units. Reduce-only exits intentionally do not call this function.
func ValidateContractOpeningMode(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, userID, symbolID int64,
	marginMode trade.MarginMode,
	positionSide trade.PositionSide,
) error {
	if svcCtx == nil || tenantID <= 0 || userID <= 0 || symbolID <= 0 {
		return errors.New("invalid contract opening mode identity")
	}
	positionMode, err := ContractPositionMode(positionSide)
	if err != nil {
		return err
	}

	var preference *models.TContractUserConfig
	for _, configuredSymbolID := range []int64{symbolID, 0} {
		item, err := svcCtx.ContractUserConfigModel.FindOneByTenantIdUserIdSymbolId(
			ctx, tenantID, userID, configuredSymbolID,
		)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if item != nil {
			preference = item
			break
		}
	}
	if err := validateContractOpeningPreference(
		preference, marginMode, positionMode,
	); err != nil {
		return err
	}

	incompatiblePositions, err := svcCtx.ContractPositionModel.CountActiveIncompatibleMode(
		ctx, tenantID, userID, symbolID, int64(marginMode), int64(positionMode),
	)
	if err != nil {
		return err
	}
	if incompatiblePositions > 0 {
		return errors.New("close incompatible positions before changing contract position or margin mode")
	}

	incompatibleOrders, err := svcCtx.TradeOrderModel.CountActiveIncompatibleContractMode(
		ctx, tenantID, userID, symbolID, int64(marginMode), int64(positionMode),
	)
	if err != nil {
		return err
	}
	if incompatibleOrders > 0 {
		return errors.New("cancel incompatible orders before changing contract position or margin mode")
	}
	return nil
}

// ContractPositionMode maps an order/fill position side to the persisted
// position mode. A one-way NET order is projected into a physical LONG/SHORT
// row, so the physical row's position_side must never be used for this choice.
func ContractPositionMode(positionSide trade.PositionSide) (trade.PositionMode, error) {
	switch positionSide {
	case trade.PositionSide_POSITION_SIDE_NET:
		return trade.PositionMode_POSITION_MODE_ONE_WAY, nil
	case trade.PositionSide_POSITION_SIDE_LONG, trade.PositionSide_POSITION_SIDE_SHORT:
		return trade.PositionMode_POSITION_MODE_HEDGE, nil
	default:
		return trade.PositionMode_POSITION_MODE_UNKNOWN, errors.New("invalid contract position side")
	}
}

func validateContractOpeningPreference(
	preference *models.TContractUserConfig,
	marginMode trade.MarginMode,
	positionMode trade.PositionMode,
) error {
	if preference == nil {
		return nil
	}
	if preference.MarginMode != int64(marginMode) {
		return errors.New("order margin mode differs from the configured user mode")
	}
	if preference.PositionMode != int64(positionMode) {
		return errors.New("order position mode differs from the configured user mode")
	}
	return nil
}
