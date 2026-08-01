package applogic

import (
	"context"
	"errors"
	"fmt"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type portfolioOrderMarginResult struct {
	margin        decimal.Decimal
	configID      int64
	configVersion int64
}

func calculatePortfolioOrderMargin(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	candidate *models.TOptionOrder,
	contract *models.TOptionContract,
	now int64,
) (portfolioOrderMarginResult, error) {
	if candidate == nil || contract == nil ||
		contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
		candidate.Side != int64(common.Side_SIDE_SELL) {
		return portfolioOrderMarginResult{}, nil
	}
	riskAccountModel := models.NewTOptionRiskAccountModel(conn, svcCtx.Config.CacheRedis)
	if _, err := riskAccountModel.EnsureAndFindOneForUpdate(
		ctx, candidate.TenantId, candidate.UserId, 0, contract.SettleCoin, now,
	); err != nil {
		return portfolioOrderMarginResult{}, err
	}
	configModel := models.NewTOptionPortfolioRiskConfigModel(conn, svcCtx.Config.CacheRedis)
	configItem, err := configModel.FindActive(ctx, candidate.TenantId, contract.SettleCoin, now)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return portfolioOrderMarginResult{}, errors.New("no approved active portfolio risk config")
		}
		return portfolioOrderMarginResult{}, err
	}
	config, err := optionrisk.PortfolioConfigFromModel(configItem)
	if err != nil {
		return portfolioOrderMarginResult{}, fmt.Errorf("invalid active portfolio risk config %d: %w", configItem.Id, err)
	}

	legs, err := loadPortfolioLegs(ctx, svcCtx, conn, candidate, contract.SettleCoin, now)
	if err != nil {
		return portfolioOrderMarginResult{}, err
	}
	before, err := optionrisk.EvaluatePortfolio(portfolioLegSlice(legs), false, config)
	if err != nil {
		return portfolioOrderMarginResult{}, err
	}
	leg, err := ensurePortfolioLeg(ctx, svcCtx, conn, legs, contract, now)
	if err != nil {
		return portfolioOrderMarginResult{}, err
	}
	if candidate.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_OPEN) {
		leg.ShortQuantity = leg.ShortQuantity.Add(candidate.Qty)
	} else {
		if leg.LongQuantity.LessThan(candidate.Qty) {
			return portfolioOrderMarginResult{}, errors.New("portfolio close order exceeds long position")
		}
		leg.LongQuantity = leg.LongQuantity.Sub(candidate.Qty)
	}
	legs[contract.Id] = leg
	after, err := optionrisk.EvaluatePortfolio(portfolioLegSlice(legs), false, config)
	if err != nil {
		return portfolioOrderMarginResult{}, err
	}
	return portfolioOrderMarginResult{
		margin:        decimal.Max(after.Requirement.Sub(before.Requirement), decimal.Zero).Round(16),
		configID:      configItem.Id,
		configVersion: configItem.Version,
	}, nil
}

func loadPortfolioLegs(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	candidate *models.TOptionOrder,
	settleCoin string,
	now int64,
) (map[int64]optionrisk.PortfolioLeg, error) {
	positionModel := models.NewTOptionPositionModel(conn, svcCtx.Config.CacheRedis)
	orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
	contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
	legs := make(map[int64]optionrisk.PortfolioLeg)

	cursor := int64(0)
	for {
		positions, _, err := positionModel.FindPage(ctx, models.OptionPositionPageFilter{
			TenantId: candidate.TenantId, UserId: candidate.UserId,
			Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		}, cursor, 100)
		if err != nil {
			return nil, err
		}
		for _, position := range positions {
			cursor = position.Id
			contract, err := contractModel.FindOne(ctx, position.ContractId)
			if err != nil {
				return nil, err
			}
			if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
				contract.SettleCoin != settleCoin {
				continue
			}
			leg, err := ensurePortfolioLeg(ctx, svcCtx, conn, legs, contract, now)
			if err != nil {
				return nil, err
			}
			if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
				leg.LongQuantity = leg.LongQuantity.Add(position.PositionQty)
			} else if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
				leg.ShortQuantity = leg.ShortQuantity.Add(position.PositionQty)
			}
			legs[contract.Id] = leg
		}
		if len(positions) < 100 {
			break
		}
	}

	orders, err := orderModel.FindPortfolioRiskOrders(
		ctx, candidate.TenantId, candidate.UserId, 0,
	)
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		contract, err := contractModel.FindOne(ctx, order.ContractId)
		if err != nil {
			return nil, err
		}
		if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
			contract.SettleCoin != settleCoin {
			continue
		}
		leg, err := ensurePortfolioLeg(ctx, svcCtx, conn, legs, contract, now)
		if err != nil {
			return nil, err
		}
		if order.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_OPEN) {
			leg.ShortQuantity = leg.ShortQuantity.Add(order.UnfilledQty)
		} else {
			leg.LongQuantity = decimal.Max(leg.LongQuantity.Sub(order.UnfilledQty), decimal.Zero)
		}
		legs[contract.Id] = leg
	}
	return legs, nil
}

func ensurePortfolioLeg(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	legs map[int64]optionrisk.PortfolioLeg,
	contract *models.TOptionContract,
	now int64,
) (optionrisk.PortfolioLeg, error) {
	if leg, ok := legs[contract.Id]; ok {
		return leg, nil
	}
	marketModel := models.NewTOptionMarketModel(conn, svcCtx.Config.CacheRedis)
	market, err := marketModel.FindOneByTenantIdContractId(ctx, contract.TenantId, contract.Id)
	if err != nil {
		return optionrisk.PortfolioLeg{}, err
	}
	if !helpers.IsRiskMarketFresh(market, now, 30) {
		return optionrisk.PortfolioLeg{}, fmt.Errorf(
			"stale portfolio market, contractId=%d underlyingSnapshotTime=%d markSnapshotTime=%d",
			contract.Id, market.UnderlyingSnapshotTime, market.MarkSnapshotTime,
		)
	}
	leg := optionrisk.PortfolioLeg{Contract: contract, Market: market}
	legs[contract.Id] = leg
	return leg, nil
}

func portfolioLegSlice(items map[int64]optionrisk.PortfolioLeg) []optionrisk.PortfolioLeg {
	result := make([]optionrisk.PortfolioLeg, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}
