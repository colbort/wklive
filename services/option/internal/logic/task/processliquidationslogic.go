package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var errPortfolioLiquidationSnapshotStale = errors.New("portfolio liquidation snapshot is stale")

type ProcessLiquidationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessLiquidationsLogic {
	return &ProcessLiquidationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行卖方强平及保险账户接管
func (l *ProcessLiquidationsLogic) ProcessLiquidations(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_liquidations", func() (*option.OptionTaskResp, error) {
		items, err := l.svcCtx.OptionLiquidationModel.FindRunnable(l.ctx, in.TenantId, 100)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if err := l.processOne(item); err != nil {
				l.Errorf("process option liquidation failed, liquidationNo=%s err=%v", item.LiquidationNo, err)
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessLiquidationsLogic) processOne(item *models.TOptionLiquidation) error {
	now := time.Now().Unix()
	if item.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING) {
		claimed, err := l.svcCtx.OptionLiquidationModel.Claim(l.ctx, item.Id, now)
		if err != nil || !claimed {
			return err
		}
	}
	current, err := l.svcCtx.OptionLiquidationModel.FindOne(l.ctx, item.Id)
	if err != nil {
		return err
	}
	instructions, err := l.svcCtx.OptionAssetInstructionModel.FindByLiquidation(l.ctx, current.TenantId, current.Id)
	if err != nil {
		return err
	}
	if len(instructions) > 0 {
		for _, instruction := range instructions {
			if instruction.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW) ||
				instruction.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED) {
				return l.markLiquidationManual(current, current.RemainingDeficit,
					"liquidation asset instruction requires manual review")
			}
		}
		return nil
	}
	waiting, err := l.cancelRiskOrders(current)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	if waiting {
		return nil
	}
	incomplete, err := l.hasLiquidationBarrier(current)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	if incomplete {
		l.Infof("option liquidation waiting for portfolio barriers, liquidationNo=%s", current.LiquidationNo)
		return nil
	}
	plan, err := l.buildLiquidationPlan(current)
	if err != nil {
		if errors.Is(err, errPortfolioLiquidationSnapshotStale) {
			reason := err.Error()
			if len(reason) > 500 {
				reason = reason[:500]
			}
			_, cancelErr := l.svcCtx.OptionLiquidationModel.CancelStalePortfolio(
				l.ctx, current.Id, time.Now().Unix(), reason,
			)
			return errors.Join(err, cancelErr)
		}
		return l.failLiquidation(current, err)
	}
	useBackstop := plan.contract.LiquidationDeficitPolicy ==
		int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP)
	insuranceCovered, err := l.coverDeficit(current, plan.deficit, plan.contract, useBackstop)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	remaining := decimal.Max(plan.deficit.Sub(insuranceCovered), decimal.Zero)
	backstopCovered := decimal.Zero
	if remaining.IsPositive() {
		if !useBackstop {
			return l.markLiquidationManual(current, remaining,
				"insurance fund insufficient; contract requires manual deficit resolution")
		}
		backstopCovered, err = l.coverPlatformBackstop(current, remaining, plan.contract)
		if err != nil {
			return l.failLiquidation(current, err)
		}
		if !backstopCovered.Equal(remaining) {
			return l.failLiquidation(current, errors.New("platform backstop did not cover the full deficit"))
		}
	}
	totalCovered := insuranceCovered.Add(backstopCovered)
	if err := validateLiquidationPlanBalance(plan, totalCovered); err != nil {
		return l.failLiquidation(current, err)
	}
	if err := l.createLiquidationInstructions(
		current, plan, insuranceCovered, backstopCovered,
	); err != nil {
		return l.failLiquidation(current, err)
	}
	return nil
}

func validateLiquidationPlanBalance(plan *optionLiquidationPlan, covered decimal.Decimal) error {
	if plan == nil || !plan.quantity.IsPositive() || !plan.takeoverCost.IsPositive() ||
		plan.fee.IsNegative() || plan.collateral.IsNegative() || covered.IsNegative() {
		return errors.New("invalid option liquidation amounts")
	}
	debit := plan.collateral.Add(covered)
	credit := plan.takeoverCost.Add(plan.fee)
	if !debit.Equal(credit) || !credit.Equal(plan.totalRequired) {
		return fmt.Errorf("unbalanced option liquidation: debit=%s credit=%s", debit, credit)
	}
	return nil
}

type optionLiquidationPlan struct {
	contract                *models.TOptionContract
	position                *models.TOptionPosition
	lots                    []*models.TOptionMarginLot
	quantity                decimal.Decimal
	takeoverCost            decimal.Decimal
	fee                     decimal.Decimal
	totalRequired           decimal.Decimal
	collateral              decimal.Decimal
	sourceCollateral        decimal.Decimal
	residualCollateralFloor decimal.Decimal
	deficit                 decimal.Decimal
}

type optionLiquidationLotAllocation struct {
	lot      *models.TOptionMarginLot
	quantity decimal.Decimal
	margin   decimal.Decimal
}

func allocateIsolatedLiquidationLots(
	lots []*models.TOptionMarginLot,
	quantity decimal.Decimal,
) ([]optionLiquidationLotAllocation, decimal.Decimal, error) {
	if !quantity.IsPositive() {
		return nil, decimal.Zero, errors.New("option liquidation quantity must be positive")
	}
	quantityLeft := quantity
	allocatedMargin := decimal.Zero
	allocations := make([]optionLiquidationLotAllocation, 0, len(lots))
	for _, lot := range lots {
		if !quantityLeft.IsPositive() {
			break
		}
		if lot == nil || !lot.RemainingQuantity.IsPositive() {
			continue
		}
		if lot.PendingMargin.IsPositive() {
			return nil, decimal.Zero, errors.New("option liquidation margin lot still has pending amount")
		}
		if lot.RemainingMargin.IsNegative() {
			return nil, decimal.Zero, errors.New("option liquidation margin lot has negative remaining margin")
		}
		takeQty := decimal.Min(quantityLeft, lot.RemainingQuantity)
		takeMargin := lot.RemainingMargin
		if takeQty.LessThan(lot.RemainingQuantity) {
			takeMargin = lot.RemainingMargin.Mul(takeQty).Div(lot.RemainingQuantity).Round(16)
		}
		if takeMargin.IsNegative() || takeMargin.GreaterThan(lot.RemainingMargin) {
			return nil, decimal.Zero, errors.New("invalid proportional option liquidation margin allocation")
		}
		allocations = append(allocations, optionLiquidationLotAllocation{
			lot: lot, quantity: takeQty, margin: takeMargin,
		})
		allocatedMargin = allocatedMargin.Add(takeMargin)
		quantityLeft = quantityLeft.Sub(takeQty)
	}
	if quantityLeft.IsPositive() {
		return nil, decimal.Zero, fmt.Errorf(
			"insufficient option margin lot quantity for liquidation: %s", quantityLeft,
		)
	}
	return allocations, allocatedMargin.Round(16), nil
}

func (l *ProcessLiquidationsLogic) buildLiquidationPlan(liq *models.TOptionLiquidation) (*optionLiquidationPlan, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, liq.ContractId)
	if err != nil {
		return nil, err
	}
	if (contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) &&
		contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)) ||
		contract.InsuranceUserId <= 0 || contract.InsuranceAccountId <= 0 {
		return nil, errors.New("option liquidation insurance takeover account is not configured")
	}
	position, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, liq.PositionId)
	if err != nil {
		return nil, err
	}
	if position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
		position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) ||
		!position.PositionQty.IsPositive() {
		return nil, errors.New("option liquidation position is no longer active")
	}
	if !liq.Quantity.IsPositive() || liq.Quantity.GreaterThan(position.PositionQty) {
		return nil, errors.New("invalid option liquidation quantity")
	}
	quantity := liq.Quantity
	if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
		if !quantity.Equal(position.PositionQty) {
			return nil, errors.New("partial portfolio liquidation is not supported")
		}
		return l.buildPortfolioLiquidationPlan(liq, contract, position)
	}
	var lots []*models.TOptionMarginLot
	lots, err = l.svcCtx.OptionMarginLotModel.FindActiveByPosition(l.ctx, liq.TenantId, position.Id)
	if err != nil {
		return nil, err
	}
	for _, lot := range lots {
		if lot.CollateralCoin != contract.SettleCoin {
			return nil, errors.New("cross-coin option liquidation requires an approved conversion model")
		}
	}
	_, collateralAvailable, err := allocateIsolatedLiquidationLots(lots, quantity)
	if err != nil {
		return nil, err
	}
	takeoverCost := liq.MarkPrice.Mul(quantity).Mul(optionMultiplier(contract)).Round(16)
	fee := takeoverCost.Mul(contract.LiquidationFeeRate).Round(16)
	total := takeoverCost.Add(fee)
	collateral := decimal.Min(collateralAvailable, total)
	return &optionLiquidationPlan{
		contract: contract, position: position, lots: lots, quantity: quantity,
		takeoverCost: takeoverCost, fee: fee, totalRequired: total,
		collateral: collateral, sourceCollateral: collateralAvailable,
		deficit: decimal.Max(total.Sub(collateral), decimal.Zero),
	}, nil
}

func (l *ProcessLiquidationsLogic) buildPortfolioLiquidationPlan(
	liq *models.TOptionLiquidation,
	contract *models.TOptionContract,
	position *models.TOptionPosition,
) (*optionLiquidationPlan, error) {
	stale := func(reason string) error {
		return fmt.Errorf("%w: %s", errPortfolioLiquidationSnapshotStale, reason)
	}
	if liq.LiquidationScope != int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET) ||
		liq.AccountId != 0 || liq.PortfolioRiskConfigId <= 0 || liq.PortfolioRiskConfigVersion <= 0 {
		return nil, errors.New("invalid portfolio liquidation audit snapshot")
	}
	if !liq.PortfolioMaintenanceBefore.GreaterThan(liq.PortfolioMaintenanceAfter) ||
		liq.PortfolioInitialAfter.IsNegative() ||
		liq.PortfolioCollateralBefore.LessThan(liq.PortfolioCollateralAfter) ||
		liq.PortfolioCollateralAfter.LessThan(liq.PortfolioInitialAfter) {
		return nil, errors.New("invalid portfolio liquidation collateral proof")
	}
	now := time.Now().Unix()
	active, err := l.svcCtx.OptionPortfolioRiskConfigModel.FindActive(
		l.ctx, liq.TenantId, contract.SettleCoin, now,
	)
	if err != nil {
		return nil, err
	}
	if active.Id != liq.PortfolioRiskConfigId || active.Version != liq.PortfolioRiskConfigVersion {
		return nil, stale("approved risk config changed")
	}
	config, err := optionrisk.PortfolioConfigFromModel(active)
	if err != nil {
		return nil, err
	}
	groups, err := NewProcessRiskAccountsLogic(l.ctx, l.svcCtx).collectRiskGroups(liq.TenantId)
	if err != nil {
		return nil, err
	}
	var group *optionRiskGroup
	for _, candidateGroup := range groups {
		if candidateGroup.key.userID == liq.UserId && candidateGroup.key.coin == contract.SettleCoin {
			group = candidateGroup
			break
		}
	}
	if group == nil || group.scanErr != nil {
		return nil, stale("wallet portfolio or fresh market set changed")
	}
	legs := make(map[int64]optionrisk.PortfolioLeg)
	for _, item := range group.positions {
		if item.contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
			continue
		}
		leg := legs[item.contract.Id]
		leg.Contract = item.contract
		leg.Market = item.market
		if item.position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			leg.LongQuantity = leg.LongQuantity.Add(item.position.PositionQty)
		} else {
			leg.ShortQuantity = leg.ShortQuantity.Add(item.position.PositionQty)
		}
		legs[item.contract.Id] = leg
	}
	legList := make([]optionrisk.PortfolioLeg, 0, len(legs))
	for _, leg := range legs {
		legList = append(legList, leg)
	}
	maintenanceBefore, err := optionrisk.EvaluatePortfolio(legList, true, config)
	if err != nil {
		return nil, err
	}
	var selected *optionRiskPosition
	isolatedMaintenance := decimal.Zero
	netOptionValue := decimal.Zero
	for i := range group.positions {
		item := &group.positions[i]
		value := item.market.MarkPrice.Mul(item.position.PositionQty).
			Mul(optionMultiplier(item.contract)).Round(16)
		if item.position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			netOptionValue = netOptionValue.Add(value)
		} else {
			netOptionValue = netOptionValue.Sub(value)
			if item.contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) {
				isolatedMaintenance = isolatedMaintenance.Add(optionTaskSellerMargin(
					item.contract, item.market.UnderlyingPrice, item.market.MarkPrice, item.position.PositionQty,
				))
			}
		}
		if item.position.Id == position.Id {
			selected = item
		}
	}
	if selected == nil || selected.position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) ||
		selected.contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
		!selected.position.PositionQty.Equal(liq.Quantity) {
		return nil, stale("selected portfolio position changed")
	}
	residualLegs := make([]optionrisk.PortfolioLeg, 0, len(legs))
	for contractID, original := range legs {
		leg := original
		if contractID == selected.contract.Id {
			leg.ShortQuantity = leg.ShortQuantity.Sub(selected.position.PositionQty)
			if leg.ShortQuantity.IsNegative() {
				return nil, stale("selected portfolio quantity exceeds current short leg")
			}
		}
		residualLegs = append(residualLegs, leg)
	}
	initialAfter, err := optionrisk.EvaluatePortfolio(residualLegs, false, config)
	if err != nil {
		return nil, err
	}
	maintenanceAfter, err := optionrisk.EvaluatePortfolio(residualLegs, true, config)
	if err != nil {
		return nil, err
	}
	if !maintenanceBefore.Requirement.GreaterThan(maintenanceAfter.Requirement) {
		return nil, stale("selected position no longer reduces current portfolio risk")
	}
	balance, err := l.svcCtx.AssetClient.GetAssetBalance(l.ctx, &asset.GetUserAssetDetailReq{
		TenantId: liq.TenantId, UserId: liq.UserId,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: contract.SettleCoin,
	})
	if err != nil {
		return nil, err
	}
	if balance.GetBase() == nil || balance.GetBase().GetCode() != 200 || balance.GetData() == nil {
		return nil, errors.New("get portfolio liquidation wallet balance rejected")
	}
	totalAsset, err := decimal.NewFromString(balance.GetData().GetTotalAmount())
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio liquidation wallet total: %w", err)
	}
	currentEquity := optionRiskEquity(totalAsset, netOptionValue)
	currentMaintenance := isolatedMaintenance.Add(maintenanceBefore.Requirement).Round(16)
	if currentEquity.GreaterThan(currentMaintenance) {
		return nil, stale("wallet recovered above current maintenance requirement")
	}
	pending, err := l.svcCtx.OptionMarginLotModel.HasPendingPortfolioByWallet(
		l.ctx, liq.TenantId, liq.UserId, contract.SettleCoin,
	)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, errors.New("portfolio liquidation margin lot still has pending amount")
	}
	lots, err := l.svcCtx.OptionMarginLotModel.FindPortfolioActiveByAccount(
		l.ctx, liq.TenantId, liq.UserId, 0, contract.SettleCoin,
	)
	if err != nil {
		return nil, err
	}
	collateralBefore := decimal.Zero
	for _, lot := range lots {
		collateralBefore = collateralBefore.Add(lot.RemainingMargin)
	}
	residualFloor := decimal.Max(liq.PortfolioInitialAfter, initialAfter.Requirement).Round(16)
	if collateralBefore.LessThan(residualFloor) {
		return nil, stale("current collateral cannot protect residual portfolio")
	}
	takeoverCost := liq.MarkPrice.Mul(liq.Quantity).Mul(optionMultiplier(contract)).Round(16)
	fee := takeoverCost.Mul(contract.LiquidationFeeRate).Round(16)
	collateral := decimal.Min(
		decimal.Max(collateralBefore.Sub(residualFloor), decimal.Zero),
		takeoverCost.Add(fee),
	)
	collateralAfter := collateralBefore.Sub(collateral).Round(16)
	if collateralAfter.LessThan(residualFloor) || !fee.Equal(liq.LiquidationFee) {
		return nil, stale("portfolio collateral proof or liquidation price changed")
	}
	return &optionLiquidationPlan{
		contract: contract, position: position, lots: lots, quantity: liq.Quantity,
		takeoverCost: takeoverCost, fee: fee, totalRequired: takeoverCost.Add(fee),
		collateral: collateral, residualCollateralFloor: residualFloor,
		deficit: decimal.Max(takeoverCost.Add(fee).Sub(collateral), decimal.Zero),
	}, nil
}

func (l *ProcessLiquidationsLogic) coverDeficit(
	liq *models.TOptionLiquidation,
	deficit decimal.Decimal,
	contract *models.TOptionContract,
	retainPartial bool,
) (decimal.Decimal, error) {
	if !deficit.IsPositive() {
		return decimal.Zero, nil
	}
	flowNo := liquidationInsuranceFlowNo(liq, retainPartial)
	flow, err := l.svcCtx.OptionInsuranceFundFlowModel.FindOneByTenantIdFlowNo(l.ctx, liq.TenantId, flowNo)
	if err == nil {
		return flow.Amount, nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return decimal.Zero, err
	}
	resp, err := l.svcCtx.AssetClient.CoverInsuranceDeficit(l.ctx, &asset.CoverInsuranceDeficitReq{
		TenantId: liq.TenantId, Coin: contract.SettleCoin,
		RequestedAmount: deficit.String(), LiquidationId: liq.Id,
		LiquidationNo: flowNo, Remark: "option liquidation insurance takeover",
	})
	if err != nil {
		return decimal.Zero, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return decimal.Zero, errors.New("option insurance fund cover rejected")
	}
	covered, err := decimal.NewFromString(resp.GetCoveredAmount())
	if err != nil || covered.IsNegative() || covered.GreaterThan(deficit) {
		return decimal.Zero, errors.New("invalid option insurance coverage")
	}
	remaining, err := decimal.NewFromString(resp.GetRemainingAmount())
	if err != nil || !covered.Add(remaining).Equal(deficit) {
		return decimal.Zero, errors.New("invalid option insurance remainder")
	}
	if remaining.IsPositive() && covered.IsPositive() && !retainPartial {
		reverse, reverseErr := l.svcCtx.AssetClient.ReverseInsuranceCover(l.ctx, &asset.ReverseInsuranceCoverReq{
			TenantId: liq.TenantId, LiquidationNo: flowNo,
			ReversalNo: flowNo + "-REVERSE", Remark: "reverse partial option insurance cover",
		})
		if reverseErr != nil || reverse == nil || reverse.GetBase() == nil || reverse.GetBase().GetCode() != 200 {
			return decimal.Zero, errors.New("partial option insurance cover reversal failed")
		}
		return decimal.Zero, nil
	}
	if covered.IsPositive() {
		_, err = l.svcCtx.OptionInsuranceFundFlowModel.Insert(l.ctx, &models.TOptionInsuranceFundFlow{
			TenantId: liq.TenantId, FlowNo: flowNo, ContractId: liq.ContractId,
			LiquidationId: liq.Id,
			FlowType:      int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER),
			Coin:          contract.SettleCoin, Amount: covered, AssetFlowNo: flowNo,
			CreateTimes: time.Now().Unix(),
		})
		if err != nil {
			return decimal.Zero, err
		}
	}
	return covered, nil
}

func (l *ProcessLiquidationsLogic) coverPlatformBackstop(
	liq *models.TOptionLiquidation,
	deficit decimal.Decimal,
	contract *models.TOptionContract,
) (decimal.Decimal, error) {
	if !deficit.IsPositive() {
		return decimal.Zero, nil
	}
	resp, err := l.svcCtx.AssetClient.CoverPlatformBackstopDeficit(
		l.ctx,
		&asset.CoverPlatformBackstopDeficitReq{
			TenantId: liq.TenantId, Coin: contract.SettleCoin,
			RequestedAmount: deficit.String(), LiquidationId: liq.Id,
			LiquidationNo: liquidationBackstopFlowNo(liq),
			Remark:        "option liquidation platform backstop liability",
		},
	)
	if err != nil {
		return decimal.Zero, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return decimal.Zero, errors.New("option platform backstop cover rejected")
	}
	covered, err := decimal.NewFromString(resp.GetCoveredAmount())
	if err != nil || !covered.Equal(deficit) {
		return decimal.Zero, errors.New("invalid option platform backstop coverage")
	}
	return covered, nil
}

func liquidationInsuranceFlowNo(liq *models.TOptionLiquidation, retainPartial bool) string {
	if retainPartial {
		return liq.LiquidationNo + "-INSURANCE-BACKSTOP"
	}
	return fmt.Sprintf("%s-INSURANCE-A%d", liq.LiquidationNo, liq.InsuranceAttempt)
}

func liquidationBackstopFlowNo(liq *models.TOptionLiquidation) string {
	return liq.LiquidationNo + "-BACKSTOP"
}

func (l *ProcessLiquidationsLogic) createLiquidationInstructions(
	liq *models.TOptionLiquidation,
	plan *optionLiquidationPlan,
	insuranceCovered decimal.Decimal,
	backstopCovered decimal.Decimal,
) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		liquidationModel := models.NewTOptionLiquidationModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := liquidationModel.FindOneForUpdate(ctx, liq.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING) {
			return nil
		}
		existing, err := instructionModel.FindByLiquidation(ctx, current.TenantId, current.Id)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return nil
		}
		position, err := positionModel.FindOneForUpdate(ctx, plan.position.Id)
		if err != nil {
			return err
		}
		if position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			position.PositionQty.LessThan(plan.quantity) {
			return errors.New("option liquidation position changed during preparation")
		}
		var lots []*models.TOptionMarginLot
		var isolatedAllocations []optionLiquidationLotAllocation
		if plan.contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
			lots, err = marginLotModel.FindPortfolioActiveByAccount(
				ctx, current.TenantId, position.UserId, 0, plan.contract.SettleCoin,
			)
		} else {
			lots, err = marginLotModel.FindActiveByPosition(ctx, current.TenantId, position.Id)
			if err == nil {
				for _, lot := range lots {
					if lot.CollateralCoin != plan.contract.SettleCoin {
						return errors.New("cross-coin option liquidation requires an approved conversion model")
					}
				}
				var allocated decimal.Decimal
				isolatedAllocations, allocated, err = allocateIsolatedLiquidationLots(lots, plan.quantity)
				if err == nil && !allocated.Equal(plan.sourceCollateral) {
					return errors.New("option liquidation collateral allocation changed during preparation")
				}
			}
		}
		if err != nil {
			return err
		}
		portfolioMode := plan.contract.SellerMarginMode ==
			int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
		if portfolioMode {
			currentPool := decimal.Zero
			for _, lot := range lots {
				if lot.PendingMargin.IsPositive() {
					return errors.New("option liquidation margin lot changed during preparation")
				}
				currentPool = currentPool.Add(lot.RemainingMargin)
			}
			if currentPool.Sub(plan.collateral).LessThan(plan.residualCollateralFloor) {
				return errors.New("option liquidation would breach residual portfolio collateral floor")
			}
		}
		now := time.Now().Unix()
		deductLeft := plan.collateral
		quantityLeft := plan.quantity
		allocationByLot := make(map[int64]optionLiquidationLotAllocation, len(isolatedAllocations))
		for _, allocation := range isolatedAllocations {
			allocationByLot[allocation.lot.Id] = allocation
		}
		for _, lot := range lots {
			if lot.PendingMargin.IsPositive() {
				return errors.New("option liquidation margin lot changed during preparation")
			}
			allocatedMargin := lot.RemainingMargin
			allocatedQuantity := decimal.Zero
			if !portfolioMode {
				allocation, ok := allocationByLot[lot.Id]
				if !ok {
					continue
				}
				allocatedMargin = allocation.margin
				allocatedQuantity = allocation.quantity
			}
			deduct := decimal.Min(allocatedMargin, deductLeft)
			release := decimal.Zero
			if !portfolioMode {
				release = allocatedMargin.Sub(deduct)
			}
			if deduct.IsPositive() {
				if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
					TenantId:      current.TenantId,
					InstructionNo: fmt.Sprintf("%s-L%d-DEDUCT", current.LiquidationNo, lot.Id),
					BizNo:         current.LiquidationNo, PositionId: position.Id, MarginLotId: lot.Id,
					LiquidationId: current.Id, UserId: position.UserId, AccountId: lot.AccountId,
					Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
					TargetBizNo: lot.FreezeBizNo, Coin: plan.contract.SettleCoin, Amount: deduct,
					StepNo: 1, CreateTimes: now, UpdateTimes: now,
				}); err != nil {
					return err
				}
				deductLeft = deductLeft.Sub(deduct)
			}
			if release.IsPositive() {
				if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
					TenantId:      current.TenantId,
					InstructionNo: fmt.Sprintf("%s-L%d-RELEASE", current.LiquidationNo, lot.Id),
					BizNo:         current.LiquidationNo, PositionId: position.Id, MarginLotId: lot.Id,
					LiquidationId: current.Id, UserId: position.UserId, AccountId: position.AccountId,
					Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
					TargetBizNo: lot.FreezeBizNo, Coin: plan.contract.SettleCoin, Amount: release,
					StepNo: 1, CreateTimes: now, UpdateTimes: now,
				}); err != nil {
					return err
				}
			}
			lot.PendingMargin = lot.PendingMargin.Add(deduct).Add(release)
			if !portfolioMode {
				lot.RemainingQuantity = decimal.Max(
					lot.RemainingQuantity.Sub(allocatedQuantity), decimal.Zero,
				)
				quantityLeft = quantityLeft.Sub(allocatedQuantity)
			}
			if deduct.IsPositive() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
			} else if release.IsPositive() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
			} else if lot.RemainingQuantity.IsZero() && lot.RemainingMargin.IsZero() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
			}
			lot.UpdateTimes = now
			if err := marginLotModel.Update(ctx, lot); err != nil {
				return err
			}
			if portfolioMode && !deductLeft.IsPositive() {
				break
			}
		}
		if deductLeft.IsPositive() {
			return errors.New("option liquidation collateral became insufficient")
		}
		if !portfolioMode && quantityLeft.IsPositive() {
			return errors.New("option liquidation margin lot quantity became insufficient")
		}
		if position.AvailableQty.LessThan(plan.quantity) {
			return errors.New("option liquidation position quantity is not available for reservation")
		}
		position.AvailableQty = position.AvailableQty.Sub(plan.quantity)
		position.FrozenQty = position.FrozenQty.Add(plan.quantity)
		position.UpdateTimes = now
		if err := positionModel.Update(ctx, position); err != nil {
			return err
		}
		if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: current.TenantId, InstructionNo: current.LiquidationNo + "-TAKEOVER-CREDIT",
			BizNo: current.LiquidationNo, PositionId: position.Id, LiquidationId: current.Id,
			UserId: plan.contract.InsuranceUserId, AccountId: plan.contract.InsuranceAccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
			Coin:   plan.contract.SettleCoin, Amount: plan.takeoverCost,
			StepNo: 2, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		if plan.fee.IsPositive() {
			if plan.contract.FeeUserId <= 0 || plan.contract.FeeAccountId <= 0 {
				return errors.New("option liquidation fee account is not configured")
			}
			if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
				TenantId: current.TenantId, InstructionNo: current.LiquidationNo + "-FEE-CREDIT",
				BizNo: current.LiquidationNo, PositionId: position.Id, LiquidationId: current.Id,
				UserId: plan.contract.FeeUserId, AccountId: plan.contract.FeeAccountId,
				Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
				Coin:   plan.contract.SettleCoin, Amount: plan.fee,
				StepNo: 2, CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: current.TenantId, InstructionNo: current.LiquidationNo + "-TAKEOVER-FREEZE",
			BizNo: current.LiquidationNo, PositionId: position.Id, LiquidationId: current.Id,
			UserId: plan.contract.InsuranceUserId, AccountId: plan.contract.InsuranceAccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
			TargetBizNo: current.LiquidationNo + "-TAKEOVER-MARGIN",
			Coin:        plan.contract.SettleCoin, Amount: plan.takeoverCost,
			StepNo: 3, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		current.Quantity = plan.quantity
		current.CollateralAmount = plan.collateral
		current.InsuranceFundAmount = insuranceCovered
		current.BackstopAmount = backstopCovered
		current.DeficitAmount = plan.deficit
		current.RemainingDeficit = decimal.Zero
		current.DeficitResolution = int64(liquidationDeficitResolution(
			plan.deficit, insuranceCovered, backstopCovered,
		))
		current.LiquidationFee = plan.fee
		current.LastErrorMsg = ""
		current.UpdateTimes = now
		return liquidationModel.Update(ctx, current)
	})
}

func liquidationDeficitResolution(
	deficit, insuranceCovered, backstopCovered decimal.Decimal,
) option.LiquidationDeficitResolution {
	if !deficit.IsPositive() {
		return option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE
	}
	if insuranceCovered.IsPositive() && backstopCovered.IsPositive() {
		return option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_INSURANCE_AND_BACKSTOP
	}
	if backstopCovered.IsPositive() {
		return option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_PLATFORM_BACKSTOP
	}
	if insuranceCovered.IsPositive() {
		return option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_INSURANCE_FUND
	}
	return option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_MANUAL_REVIEW
}

func insertLiquidationInstruction(
	ctx context.Context,
	model models.TOptionAssetInstructionModel,
	item *models.TOptionAssetInstruction,
) error {
	item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING)
	item.ReconciliationStatus = int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING)
	_, err := model.Insert(ctx, item)
	return err
}

func (l *ProcessLiquidationsLogic) cancelRiskOrders(liq *models.TOptionLiquidation) (bool, error) {
	if liq.LiquidationScope == int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET) {
		contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, liq.ContractId)
		if err != nil {
			return false, err
		}
		orders, err := l.svcCtx.OptionOrderModel.FindPortfolioRiskOrders(l.ctx, liq.TenantId, liq.UserId, 0)
		if err != nil {
			return false, err
		}
		waiting := false
		for _, order := range orders {
			orderContract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, order.ContractId)
			if err != nil {
				return false, err
			}
			if orderContract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
				orderContract.SettleCoin != contract.SettleCoin {
				continue
			}
			waiting = true
			if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) ||
				order.Status == int64(option.OrderStatus_ORDER_STATUS_PENDING) ||
				order.Status == int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
				if err := cancelOptionSystemOrder(l.ctx, l.svcCtx, order.Id, "PORTFOLIO_RISK_LIQUIDATION"); err != nil {
					return false, err
				}
			}
		}
		return waiting, nil
	}
	cursor := int64(0)
	waiting := false
	for {
		orders, _, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
			TenantId: liq.TenantId, UserId: liq.UserId, AccountId: liq.AccountId,
			ContractId: liq.ContractId,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_FUNDING),
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return false, err
		}
		for _, order := range orders {
			cursor = order.Id
			waiting = true
			if err := cancelOptionSystemOrder(l.ctx, l.svcCtx, order.Id, "RISK_LIQUIDATION"); err != nil {
				return false, err
			}
		}
		if len(orders) < 100 {
			return waiting, nil
		}
	}
}

func (l *ProcessLiquidationsLogic) hasLiquidationBarrier(liq *models.TOptionLiquidation) (bool, error) {
	if liq.LiquidationScope != int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET) {
		incomplete, err := l.svcCtx.OptionOutboxModel.HasIncomplete(l.ctx, liq.TenantId, liq.ContractId)
		if err != nil || incomplete {
			return incomplete, err
		}
		return l.svcCtx.OptionAssetInstructionModel.HasIncompleteMarginForContract(
			l.ctx, liq.TenantId, liq.ContractId,
		)
	}
	sourceContract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, liq.ContractId)
	if err != nil {
		return false, err
	}
	incomplete, err := l.svcCtx.OptionOutboxModel.HasIncompletePortfolioForWallet(
		l.ctx, liq.TenantId, liq.UserId, sourceContract.SettleCoin,
	)
	if err != nil || incomplete {
		return incomplete, err
	}
	return l.svcCtx.OptionAssetInstructionModel.HasIncompleteForWallet(
		l.ctx, liq.TenantId, liq.UserId, sourceContract.SettleCoin,
	)
}

func cancelOptionSystemOrder(ctx context.Context, svcCtx *svc.ServiceContext, orderID int64, reason string) error {
	var canceled *models.TOptionOrder
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		changed, err := cancelOptionSystemOrderInTx(
			txCtx, orderModel, positionModel, instructionModel, order, reason, now,
		)
		if err != nil {
			return err
		}
		if changed {
			canceled = order
		}
		return nil
	})
	if err == nil && canceled != nil {
		applogic.PublishOptionOrderChanged(ctx, svcCtx, canceled)
	}
	return err
}

func cancelOptionSystemOrderInTx(
	ctx context.Context,
	orderModel models.TOptionOrderModel,
	positionModel models.TOptionPositionModel,
	instructionModel models.TOptionAssetInstructionModel,
	order *models.TOptionOrder,
	reason string,
	now int64,
) (bool, error) {
	switch option.OrderStatus(order.Status) {
	case option.OrderStatus_ORDER_STATUS_FUNDING,
		option.OrderStatus_ORDER_STATUS_PENDING,
		option.OrderStatus_ORDER_STATUS_PART_FILLED:
	default:
		return false, nil
	}
	cancelBeforeFreeze := false
	if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
		freeze, err := instructionModel.FindOneByTenantIdInstructionNo(ctx, order.TenantId, order.OrderNo+"-FREEZE")
		if err != nil {
			return false, err
		}
		freeze, err = instructionModel.FindOneForUpdate(ctx, freeze.Id)
		if err != nil {
			return false, err
		}
		if freeze.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) {
			freeze.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED)
			freeze.UpdateTimes = now
			if err := instructionModel.Update(ctx, freeze); err != nil {
				return false, err
			}
			cancelBeforeFreeze = true
			order.MarginAmount = decimal.Zero
		}
	}
	if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
		return false, err
	}
	order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	if order.MarginAmount.IsPositive() && !cancelBeforeFreeze {
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: order.TenantId, InstructionNo: order.OrderNo + "-LIQUIDATION-RELEASE",
			BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
			TargetBizNo: order.OrderNo, Coin: applogic.OptionOrderMarginCoin(order), Amount: order.MarginAmount,
			StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return false, err
		}
	}
	order.CancelReason = reason
	order.CancelTime = now
	order.UpdateTimes = now
	if err := orderModel.Update(ctx, order); err != nil {
		return false, err
	}
	return true, nil
}

func (l *ProcessLiquidationsLogic) failLiquidation(liq *models.TOptionLiquidation, cause error) error {
	current, err := l.svcCtx.OptionLiquidationModel.FindOne(l.ctx, liq.Id)
	if err != nil {
		return errors.Join(cause, err)
	}
	current.RetryCount++
	current.Status = int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED)
	if current.RetryCount >= 20 {
		current.Status = int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)
	}
	current.LastErrorMsg = cause.Error()
	if len(current.LastErrorMsg) > 500 {
		current.LastErrorMsg = current.LastErrorMsg[:500]
	}
	current.UpdateTimes = time.Now().Unix()
	return errors.Join(cause, l.svcCtx.OptionLiquidationModel.Update(l.ctx, current))
}

func (l *ProcessLiquidationsLogic) markLiquidationManual(
	liq *models.TOptionLiquidation,
	remaining decimal.Decimal,
	reason string,
) error {
	current, err := l.svcCtx.OptionLiquidationModel.FindOne(l.ctx, liq.Id)
	if err != nil {
		return err
	}
	current.Status = int64(option.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)
	current.RemainingDeficit = decimal.Max(remaining, decimal.Zero)
	current.DeficitResolution = int64(
		option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_MANUAL_REVIEW,
	)
	current.LastErrorMsg = reason
	current.UpdateTimes = time.Now().Unix()
	return l.svcCtx.OptionLiquidationModel.Update(l.ctx, current)
}
