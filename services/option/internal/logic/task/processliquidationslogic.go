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
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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

// 执行卖方逐仓强平及保险账户接管
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
	if err := l.cancelRiskOrders(current); err != nil {
		return l.failLiquidation(current, err)
	}
	incomplete, err := l.svcCtx.OptionOutboxModel.HasIncomplete(l.ctx, current.TenantId, current.ContractId)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	if incomplete {
		l.Infof("option liquidation waiting for contract trade events, liquidationNo=%s", current.LiquidationNo)
		return nil
	}
	incomplete, err = l.svcCtx.OptionAssetInstructionModel.HasIncompleteMarginForContract(
		l.ctx, current.TenantId, current.ContractId,
	)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	if incomplete {
		l.Infof("option liquidation waiting for margin instructions, liquidationNo=%s", current.LiquidationNo)
		return nil
	}
	plan, err := l.buildLiquidationPlan(current)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	covered, err := l.coverDeficit(current, plan.deficit, plan.contract)
	if err != nil {
		return l.failLiquidation(current, err)
	}
	if !covered.Equal(plan.deficit) {
		return l.markLiquidationManual(current, plan.deficit.Sub(covered),
			"insurance fund insufficient; takeover was not started")
	}
	if err := validateLiquidationPlanBalance(plan, covered); err != nil {
		return l.failLiquidation(current, err)
	}
	if err := l.createLiquidationInstructions(current, plan, covered); err != nil {
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
	contract      *models.TOptionContract
	position      *models.TOptionPosition
	lots          []*models.TOptionMarginLot
	quantity      decimal.Decimal
	takeoverCost  decimal.Decimal
	fee           decimal.Decimal
	totalRequired decimal.Decimal
	collateral    decimal.Decimal
	deficit       decimal.Decimal
}

func (l *ProcessLiquidationsLogic) buildLiquidationPlan(liq *models.TOptionLiquidation) (*optionLiquidationPlan, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, liq.ContractId)
	if err != nil {
		return nil, err
	}
	if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) ||
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
	quantity := decimal.Min(liq.Quantity, position.PositionQty)
	lots, err := l.svcCtx.OptionMarginLotModel.FindActiveByPosition(l.ctx, liq.TenantId, position.Id)
	if err != nil {
		return nil, err
	}
	collateralAvailable := decimal.Zero
	for _, lot := range lots {
		if lot.PendingMargin.IsPositive() {
			return nil, errors.New("option liquidation margin lot still has pending amount")
		}
		collateralAvailable = collateralAvailable.Add(lot.RemainingMargin)
	}
	takeoverCost := liq.MarkPrice.Mul(quantity).Mul(optionMultiplier(contract)).Round(16)
	fee := takeoverCost.Mul(contract.LiquidationFeeRate).Round(16)
	total := takeoverCost.Add(fee)
	collateral := decimal.Min(collateralAvailable, total)
	return &optionLiquidationPlan{
		contract: contract, position: position, lots: lots, quantity: quantity,
		takeoverCost: takeoverCost, fee: fee, totalRequired: total,
		collateral: collateral, deficit: decimal.Max(total.Sub(collateral), decimal.Zero),
	}, nil
}

func (l *ProcessLiquidationsLogic) coverDeficit(
	liq *models.TOptionLiquidation,
	deficit decimal.Decimal,
	contract *models.TOptionContract,
) (decimal.Decimal, error) {
	if !deficit.IsPositive() {
		return decimal.Zero, nil
	}
	flowNo := fmt.Sprintf("%s-INSURANCE-A%d", liq.LiquidationNo, liq.InsuranceAttempt)
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
	if remaining.IsPositive() && covered.IsPositive() {
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

func (l *ProcessLiquidationsLogic) createLiquidationInstructions(
	liq *models.TOptionLiquidation,
	plan *optionLiquidationPlan,
	covered decimal.Decimal,
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
		lots, err := marginLotModel.FindActiveByPosition(ctx, current.TenantId, position.Id)
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		deductLeft := plan.collateral
		for _, lot := range lots {
			if lot.PendingMargin.IsPositive() {
				return errors.New("option liquidation margin lot changed during preparation")
			}
			deduct := decimal.Min(lot.RemainingMargin, deductLeft)
			release := lot.RemainingMargin.Sub(deduct)
			if deduct.IsPositive() {
				if err := insertLiquidationInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
					TenantId:      current.TenantId,
					InstructionNo: fmt.Sprintf("%s-L%d-DEDUCT", current.LiquidationNo, lot.Id),
					BizNo:         current.LiquidationNo, PositionId: position.Id, MarginLotId: lot.Id,
					LiquidationId: current.Id, UserId: position.UserId, AccountId: position.AccountId,
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
			lot.PendingMargin = lot.RemainingMargin
			lot.RemainingQuantity = decimal.Zero
			if deduct.IsPositive() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
			} else {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
			}
			lot.UpdateTimes = now
			if err := marginLotModel.Update(ctx, lot); err != nil {
				return err
			}
		}
		if deductLeft.IsPositive() {
			return errors.New("option liquidation collateral became insufficient")
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
		current.InsuranceFundAmount = covered
		current.DeficitAmount = plan.deficit
		current.RemainingDeficit = decimal.Zero
		current.LiquidationFee = plan.fee
		current.LastErrorMsg = ""
		current.UpdateTimes = now
		return liquidationModel.Update(ctx, current)
	})
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

func (l *ProcessLiquidationsLogic) cancelRiskOrders(liq *models.TOptionLiquidation) error {
	cursor := int64(0)
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
			return err
		}
		for _, order := range orders {
			cursor = order.Id
			if err := cancelOptionSystemOrder(l.ctx, l.svcCtx, order.Id, "RISK_LIQUIDATION"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
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
		switch option.OrderStatus(order.Status) {
		case option.OrderStatus_ORDER_STATUS_FUNDING,
			option.OrderStatus_ORDER_STATUS_PENDING,
			option.OrderStatus_ORDER_STATUS_PART_FILLED:
		default:
			return nil
		}
		now := time.Now().Unix()
		cancelBeforeFreeze := false
		if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
			freeze, err := instructionModel.FindOneByTenantIdInstructionNo(txCtx, order.TenantId, order.OrderNo+"-FREEZE")
			if err != nil {
				return err
			}
			freeze, err = instructionModel.FindOneForUpdate(txCtx, freeze.Id)
			if err != nil {
				return err
			}
			if freeze.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) {
				freeze.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED)
				freeze.UpdateTimes = now
				if err := instructionModel.Update(txCtx, freeze); err != nil {
					return err
				}
				cancelBeforeFreeze = true
				order.MarginAmount = decimal.Zero
			}
		}
		if err := releaseClosePositionFrozenQty(txCtx, positionModel, order, order.UnfilledQty, now); err != nil {
			return err
		}
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
		if order.MarginAmount.IsPositive() && !cancelBeforeFreeze {
			order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
			if _, err := instructionModel.Insert(txCtx, &models.TOptionAssetInstruction{
				TenantId: order.TenantId, InstructionNo: order.OrderNo + "-LIQUIDATION-RELEASE",
				BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: order.OrderNo, Coin: order.FeeCoin, Amount: order.MarginAmount,
				StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		order.CancelReason = reason
		order.CancelTime = now
		order.UpdateTimes = now
		if err := orderModel.Update(txCtx, order); err != nil {
			return err
		}
		canceled = order
		return nil
	})
	if err == nil && canceled != nil {
		applogic.PublishOptionOrderChanged(ctx, svcCtx, canceled)
	}
	return err
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
	current.LastErrorMsg = reason
	current.UpdateTimes = time.Now().Unix()
	return l.svcCtx.OptionLiquidationModel.Update(l.ctx, current)
}
