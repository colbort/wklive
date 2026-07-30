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

type ProcessAssetInstructionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessAssetInstructionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessAssetInstructionsLogic {
	return &ProcessAssetInstructionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行并重试 Option 资产指令
func (l *ProcessAssetInstructionsLogic) ProcessAssetInstructions(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_asset_instructions", func() (*option.OptionTaskResp, error) {
		cursor := int64(0)
		now := time.Now().Unix()
		for {
			items, err := l.svcCtx.OptionAssetInstructionModel.FindRunnable(l.ctx, in.TenantId, now, cursor, 100)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				return helpers.OkTaskResp(), nil
			}
			for _, item := range items {
				cursor = item.Id
				if err := l.processOne(item); err != nil {
					l.Errorf("process option asset instruction failed, instructionNo=%s err=%v", item.InstructionNo, err)
				}
			}
			if len(items) < 100 {
				return helpers.OkTaskResp(), nil
			}
		}
	})
}

func (l *ProcessAssetInstructionsLogic) processOne(item *models.TOptionAssetInstruction) error {
	item.UpdateTimes = time.Now().Unix()
	claimed, err := l.svcCtx.OptionAssetInstructionModel.Claim(l.ctx, item.Id, item.UpdateTimes)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING)

	err = l.execute(item)
	if err != nil {
		if updateErr := l.markInstructionFailed(item, err); updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	if err := l.reconcile(item); err != nil {
		if issueErr := l.openAssetReconciliationIssue(item, err); issueErr != nil {
			err = errors.Join(err, issueErr)
		}
		if updateErr := l.markInstructionFailed(item, err); updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	if err := l.svcCtx.OptionReconciliationIssueModel.Resolve(
		l.ctx, item.TenantId, assetReconciliationIssueKey(item), time.Now().Unix(),
	); err != nil {
		if updateErr := l.markInstructionFailed(item, err); updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}

	item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS)
	item.NextRetryAt = 0
	item.LastErrorMsg = ""
	item.UpdateTimes = time.Now().Unix()
	if err := l.svcCtx.OptionAssetInstructionModel.Update(l.ctx, item); err != nil {
		return err
	}
	if err := l.runInstructionCompletion(item); err != nil {
		if updateErr := l.markInstructionFailed(item, err); updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	return nil
}

func assetReconciliationIssueKey(item *models.TOptionAssetInstruction) string {
	return "ASSET_FLOW:" + item.InstructionNo
}

func (l *ProcessAssetInstructionsLogic) openAssetReconciliationIssue(item *models.TOptionAssetInstruction, cause error) error {
	now := time.Now().Unix()
	return l.svcCtx.OptionReconciliationIssueModel.Open(l.ctx, &models.TOptionReconciliationIssue{
		TenantId:  item.TenantId,
		IssueKey:  assetReconciliationIssueKey(item),
		CheckType: int64(option.ReconciliationCheckType_RECONCILIATION_CHECK_TYPE_ASSET_FLOW),
		BizNo:     item.BizNo, InstructionId: item.Id,
		ExpectedValue: fmt.Sprintf("user=%d wallet=%d coin=%s amount=%s action=%d",
			item.UserId, common.WalletType_WALLET_TYPE_OPTION, item.Coin, item.Amount, item.Action),
		ActualValue:     item.AssetFlowNo,
		Detail:          cause.Error(),
		Status:          int64(option.ReconciliationIssueStatus_RECONCILIATION_ISSUE_STATUS_OPEN),
		OccurrenceCount: 1,
		CreateTimes:     now, UpdateTimes: now,
	})
}

func (l *ProcessAssetInstructionsLogic) reconcile(item *models.TOptionAssetInstruction) error {
	scene, op, err := optionInstructionAssetFacts(item.Action)
	if err != nil {
		return err
	}
	flowBizNo := item.InstructionNo
	if item.Action == int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE) {
		flowBizNo = item.TargetBizNo
	}
	resp, err := l.svcCtx.AssetClient.GetAssetFlowByBizNo(l.ctx, &asset.GetAssetFlowByBizNoReq{
		TenantId:  item.TenantId,
		BizType:   asset.BizType_BIZ_TYPE_OPTION,
		SceneType: scene,
		BizNo:     flowBizNo,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil {
		if resp != nil && resp.Base != nil {
			return fmt.Errorf("asset reconciliation lookup rejected: code=%d msg=%s", resp.Base.Code, resp.Base.Msg)
		}
		return errors.New("asset reconciliation lookup returned empty response")
	}
	flowAmount, err := decimal.NewFromString(resp.Data.ChangeAmount)
	if err != nil {
		return fmt.Errorf("invalid asset reconciliation amount: %w", err)
	}
	matched := resp.Data.TenantId == item.TenantId &&
		resp.Data.UserId == item.UserId &&
		resp.Data.WalletType == common.WalletType_WALLET_TYPE_OPTION &&
		resp.Data.Coin == item.Coin &&
		resp.Data.BizType == asset.BizType_BIZ_TYPE_OPTION &&
		resp.Data.SceneType == scene &&
		resp.Data.OpType == op &&
		flowAmount.Equal(item.Amount)
	if !matched {
		item.ReconciliationStatus = int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MISMATCH)
		item.AssetFlowNo = resp.Data.FlowNo
		item.ReconciledAt = time.Now().Unix()
		return fmt.Errorf("asset reconciliation mismatch: instructionNo=%s flowNo=%s", item.InstructionNo, resp.Data.FlowNo)
	}
	if err := l.syncAccountMirror(item, resp.Data); err != nil {
		return err
	}
	item.AssetFlowNo = resp.Data.FlowNo
	item.ReconciliationStatus = int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED)
	item.ReconciledAt = time.Now().Unix()
	return nil
}

func (l *ProcessAssetInstructionsLogic) syncAccountMirror(item *models.TOptionAssetInstruction, flow *asset.AssetFlow) error {
	total, err := decimal.NewFromString(flow.AfterTotalAmount)
	if err != nil {
		return fmt.Errorf("invalid Asset total amount: %w", err)
	}
	available, err := decimal.NewFromString(flow.AfterAvailableAmount)
	if err != nil {
		return fmt.Errorf("invalid Asset available amount: %w", err)
	}
	frozen, err := decimal.NewFromString(flow.AfterFrozenAmount)
	if err != nil {
		return fmt.Errorf("invalid Asset frozen amount: %w", err)
	}
	beforeTotal, err := decimal.NewFromString(flow.BeforeTotalAmount)
	if err != nil {
		return fmt.Errorf("invalid Asset before total amount: %w", err)
	}
	now := time.Now().Unix()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accountModel := models.NewTOptionAccountModel(conn, l.svcCtx.Config.CacheRedis)
		billModel := models.NewTOptionBillModel(conn, l.svcCtx.Config.CacheRedis)
		account, err := accountModel.FindOneByTenantIdUserIdAccountIdMarginCoin(ctx, item.TenantId, item.UserId, item.AccountId, item.Coin)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if errors.Is(err, models.ErrNotFound) {
			account = &models.TOptionAccount{
				TenantId: item.TenantId, UserId: item.UserId, AccountId: item.AccountId,
				MarginCoin: item.Coin, Status: int64(option.AccountStatus_ACCOUNT_STATUS_NORMAL),
				CreateTimes: now,
			}
			account.Balance, account.AvailableBalance, account.FrozenBalance = total, available, frozen
			account.UpdateTimes = now
			if _, err := accountModel.Insert(ctx, account); err != nil {
				return err
			}
		} else {
			account.Balance, account.AvailableBalance, account.FrozenBalance = total, available, frozen
			account.UpdateTimes = now
			if err := accountModel.Update(ctx, account); err != nil {
				return err
			}
		}
		if _, err := billModel.FindOneByTenantIdBizNo(ctx, item.TenantId, item.InstructionNo); err == nil {
			return nil
		} else if !errors.Is(err, models.ErrNotFound) {
			return err
		}
		_, err = billModel.Insert(ctx, &models.TOptionBill{
			TenantId: item.TenantId, UserId: item.UserId, AccountId: item.AccountId,
			BizNo: item.InstructionNo, RefType: optionInstructionBillRefType(item),
			RefId: optionInstructionRefID(item), Coin: item.Coin,
			ChangeAmount: total.Sub(beforeTotal), BalanceBefore: beforeTotal, BalanceAfter: total,
			Remark: "Asset authoritative flow " + flow.FlowNo, CreateTimes: now,
		})
		return err
	})
}

func optionInstructionBillRefType(item *models.TOptionAssetInstruction) int64 {
	if item.PositionId > 0 {
		return int64(option.BillRefType_BILL_REF_TYPE_SETTLEMENT)
	}
	if item.TradeId > 0 {
		return int64(option.BillRefType_BILL_REF_TYPE_TRADE)
	}
	switch option.AssetInstructionAction(item.Action) {
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE:
		return int64(option.BillRefType_BILL_REF_TYPE_ORDER)
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN:
		return int64(option.BillRefType_BILL_REF_TYPE_CANCEL)
	default:
		return int64(option.BillRefType_BILL_REF_TYPE_UNKNOWN)
	}
}

func optionInstructionRefID(item *models.TOptionAssetInstruction) int64 {
	if item.PositionId > 0 {
		return item.PositionId
	}
	if item.TradeId > 0 {
		return item.TradeId
	}
	return item.OrderId
}

func optionInstructionAssetFacts(action int64) (asset.SceneType, asset.AssetOpType, error) {
	switch option.AssetInstructionAction(action) {
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE:
		return asset.SceneType_SCENE_TYPE_PLACE_ORDER, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, nil
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN:
		return asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT, nil
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN:
		return asset.SceneType_SCENE_TYPE_CANCEL_ORDER, asset.AssetOpType_ASSET_OP_TYPE_UNFREEZE, nil
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE:
		return asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_ADD, nil
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE:
		return asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_SUB, nil
	default:
		return asset.SceneType_SCENE_TYPE_UNKNOWN, asset.AssetOpType_ASSET_OP_TYPE_UNKNOWN,
			fmt.Errorf("unsupported option asset instruction action: %d", action)
	}
}

func (l *ProcessAssetInstructionsLogic) runInstructionCompletion(item *models.TOptionAssetInstruction) error {
	if err := l.completeMarginLotTransition(item); err != nil {
		return err
	}
	if err := l.completeOrderTransition(item); err != nil {
		return err
	}
	if err := l.completeFundingTransition(item); err != nil {
		return err
	}
	if err := l.completeExerciseTransition(item); err != nil {
		return err
	}
	if err := l.completeSettlementTransition(item); err != nil {
		return err
	}
	return l.completeLiquidationTransition(item)
}

func (l *ProcessAssetInstructionsLogic) completeExerciseTransition(item *models.TOptionAssetInstruction) error {
	if item.BizNo == "" {
		return nil
	}
	return completeExerciseIfReady(l.ctx, l.svcCtx, item.TenantId, item.BizNo)
}

func completeExerciseIfReady(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantId int64,
	exerciseNo string,
) error {
	exercise, err := svcCtx.OptionExerciseModel.FindOneByTenantIdExerciseNo(ctx, tenantId, exerciseNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		return err
	}
	if exercise.Status == int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) {
		return nil
	}
	instructions, err := svcCtx.OptionAssetInstructionModel.FindByBizNo(ctx, tenantId, exerciseNo)
	if err != nil {
		return err
	}
	if len(instructions) == 0 {
		return nil
	}
	for _, instruction := range instructions {
		if instruction.Status != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
	}
	return svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exerciseModel := models.NewTOptionExerciseModel(conn, svcCtx.Config.CacheRedis)
		assignmentModel := models.NewTOptionExerciseAssignmentModel(conn, svcCtx.Config.CacheRedis)
		current, err := exerciseModel.FindOneForUpdate(ctx, exercise.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) {
			return nil
		}
		if current.Status != int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING) {
			return errors.New("option exercise is not pending")
		}
		now := time.Now().Unix()
		assignments, err := assignmentModel.FindByExercise(ctx, current.TenantId, current.Id)
		if err != nil {
			return err
		}
		if len(assignments) == 0 {
			return errors.New("option exercise has no short assignments")
		}
		for _, assignment := range assignments {
			assignment.Status = int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_DONE)
			assignment.UpdateTimes = now
			if err := assignmentModel.Update(ctx, assignment); err != nil {
				return err
			}
		}
		current.Status = int64(option.ExerciseStatus_EXERCISE_STATUS_DONE)
		current.FinishTime = now
		current.UpdateTimes = now
		return exerciseModel.Update(ctx, current)
	})
}

func (l *ProcessAssetInstructionsLogic) completeLiquidationTransition(item *models.TOptionAssetInstruction) error {
	if item.LiquidationId == 0 {
		return nil
	}
	instructions, err := l.svcCtx.OptionAssetInstructionModel.FindByLiquidation(
		l.ctx, item.TenantId, item.LiquidationId,
	)
	if err != nil {
		return err
	}
	if len(instructions) == 0 {
		return nil
	}
	for _, instruction := range instructions {
		if instruction.Status != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		liquidationModel := models.NewTOptionLiquidationModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		flowModel := models.NewTOptionInsuranceFundFlowModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		liq, err := liquidationModel.FindOneForUpdate(ctx, item.LiquidationId)
		if err != nil {
			return err
		}
		if liq.Status == int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) {
			return nil
		}
		if liq.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING) {
			return errors.New("option liquidation is not executable")
		}
		contract, err := contractModel.FindOne(ctx, liq.ContractId)
		if err != nil {
			return err
		}
		source, err := positionModel.FindOneForUpdate(ctx, liq.PositionId)
		if err != nil {
			return err
		}
		if source.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			source.PositionQty.LessThan(liq.Quantity) {
			return errors.New("option liquidation source position changed before takeover")
		}
		now := time.Now().Unix()
		multiplier := optionMultiplier(contract)
		realized := source.OpenAvgPrice.Sub(liq.MarkPrice).Mul(liq.Quantity).Mul(multiplier).Round(16)
		source.RealizedPnl = source.RealizedPnl.Add(realized)
		source.PositionQty = decimal.Max(source.PositionQty.Sub(liq.Quantity), decimal.Zero)
		source.AvailableQty = decimal.Min(source.AvailableQty, source.PositionQty)
		source.FrozenQty = decimal.Max(source.PositionQty.Sub(source.AvailableQty), decimal.Zero)
		source.PositionValue = liq.MarkPrice.Mul(source.PositionQty).Mul(multiplier).Round(16)
		source.MarkPrice = liq.MarkPrice
		source.UnrealizedPnl = source.OpenAvgPrice.Sub(liq.MarkPrice).
			Mul(source.PositionQty).Mul(multiplier).Round(16)
		source.LastCalcTime = now
		source.UpdateTimes = now
		if source.PositionQty.IsZero() {
			source.AvailableQty = decimal.Zero
			source.FrozenQty = decimal.Zero
			source.PositionValue = decimal.Zero
			source.MarginAmount = decimal.Zero
			source.MaintenanceMargin = decimal.Zero
			source.UnrealizedPnl = decimal.Zero
			source.Status = int64(option.PositionStatus_POSITION_STATUS_CLOSED)
		}
		if err := positionModel.Update(ctx, source); err != nil {
			return err
		}
		takeover, err := positionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
			ctx, liq.TenantId, contract.InsuranceUserId, contract.InsuranceAccountId,
			contract.Id, int64(common.PositionSide_POSITION_SIDE_SHORT),
		)
		takeoverMargin := liq.MarkPrice.Mul(liq.Quantity).Mul(multiplier).Round(16)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if errors.Is(err, models.ErrNotFound) {
			takeover = &models.TOptionPosition{
				TenantId: liq.TenantId, UserId: contract.InsuranceUserId,
				AccountId: contract.InsuranceAccountId, ContractId: contract.Id,
				UnderlyingSymbol: contract.UnderlyingSymbol,
				Side:             int64(common.PositionSide_POSITION_SIDE_SHORT),
				PositionQty:      liq.Quantity, AvailableQty: liq.Quantity,
				OpenAvgPrice: liq.MarkPrice, MarkPrice: liq.MarkPrice,
				PositionValue: takeoverMargin, MarginAmount: takeoverMargin,
				Status:       int64(option.PositionStatus_POSITION_STATUS_HOLDING),
				LastCalcTime: now, CreateTimes: now, UpdateTimes: now,
			}
			result, err := positionModel.Insert(ctx, takeover)
			if err != nil {
				return err
			}
			takeover.Id, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			nextQty := takeover.PositionQty.Add(liq.Quantity)
			takeover.OpenAvgPrice = takeover.OpenAvgPrice.Mul(takeover.PositionQty).
				Add(liq.MarkPrice.Mul(liq.Quantity)).Div(nextQty)
			takeover.PositionQty = nextQty
			takeover.AvailableQty = takeover.AvailableQty.Add(liq.Quantity)
			takeover.MarkPrice = liq.MarkPrice
			takeover.PositionValue = liq.MarkPrice.Mul(nextQty).Mul(multiplier).Round(16)
			takeover.MarginAmount = takeover.MarginAmount.Add(takeoverMargin)
			takeover.UnrealizedPnl = takeover.OpenAvgPrice.Sub(liq.MarkPrice).Mul(nextQty).Mul(multiplier).Round(16)
			takeover.Status = int64(option.PositionStatus_POSITION_STATUS_HOLDING)
			takeover.LastCalcTime = now
			takeover.UpdateTimes = now
			if err := positionModel.Update(ctx, takeover); err != nil {
				return err
			}
		}
		takeoverMarginLot, err := marginLotModel.FindOneByTenantIdTradeId(ctx, liq.TenantId, -liq.Id)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if errors.Is(err, models.ErrNotFound) {
			_, err = marginLotModel.Insert(ctx, &models.TOptionMarginLot{
				TenantId: liq.TenantId, UserId: contract.InsuranceUserId,
				AccountId: contract.InsuranceAccountId, ContractId: contract.Id,
				PositionId: takeover.Id, TradeId: -liq.Id,
				FreezeBizNo: liq.LiquidationNo + "-TAKEOVER-MARGIN",
				Quantity:    liq.Quantity, RemainingQuantity: liq.Quantity,
				InitialMargin: takeoverMargin, RemainingMargin: takeoverMargin,
				Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
				CreateTimes: now, UpdateTimes: now,
			})
			if err != nil {
				return err
			}
		} else if takeoverMarginLot.PositionId != takeover.Id {
			return errors.New("option liquidation takeover margin lot position mismatch")
		}
		if liq.LiquidationFee.IsPositive() {
			_, err := flowModel.Insert(ctx, &models.TOptionInsuranceFundFlow{
				TenantId: liq.TenantId, FlowNo: liq.LiquidationNo + "-FEE",
				ContractId: liq.ContractId, LiquidationId: liq.Id,
				FlowType: int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_LIQUIDATION_FEE),
				Coin:     contract.SettleCoin, Amount: liq.LiquidationFee,
				AssetFlowNo: liq.LiquidationNo + "-FEE-CREDIT", CreateTimes: now,
			})
			if err != nil {
				return err
			}
		}
		liq.TakeoverPositionId = takeover.Id
		liq.Status = int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE)
		liq.RemainingDeficit = decimal.Zero
		liq.LastErrorMsg = ""
		liq.CompletedAt = now
		liq.UpdateTimes = now
		return liquidationModel.Update(ctx, liq)
	})
}

func (l *ProcessAssetInstructionsLogic) completeMarginLotTransition(item *models.TOptionAssetInstruction) error {
	if item.MarginLotId == 0 ||
		(item.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN) &&
			item.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN)) {
		return nil
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		result, err := conn.ExecCtx(ctx, `INSERT IGNORE INTO t_option_margin_lot_application
(tenant_id,instruction_id,margin_lot_id,action,amount,create_times)
VALUES (?,?,?,?,?,?)`,
			item.TenantId, item.Id, item.MarginLotId, item.Action, item.Amount, time.Now().Unix(),
		)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		lot, err := marginLotModel.FindOneForUpdate(ctx, item.MarginLotId)
		if err != nil {
			return err
		}
		consumed := decimal.Min(lot.RemainingMargin, item.Amount)
		lot.RemainingMargin = decimal.Max(lot.RemainingMargin.Sub(consumed), decimal.Zero)
		lot.PendingMargin = decimal.Max(lot.PendingMargin.Sub(consumed), decimal.Zero)
		if lot.RemainingMargin.IsZero() {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
		} else if lot.PendingMargin.IsPositive() {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
		} else {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE)
		}
		lot.UpdateTimes = time.Now().Unix()
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return err
		}
		if lot.PositionId == 0 || !consumed.IsPositive() {
			return nil
		}
		position, err := positionModel.FindOne(ctx, lot.PositionId)
		if err != nil {
			return err
		}
		position.MarginAmount = decimal.Max(position.MarginAmount.Sub(consumed), decimal.Zero)
		position.UpdateTimes = lot.UpdateTimes
		return positionModel.Update(ctx, position)
	})
}

func (l *ProcessAssetInstructionsLogic) markInstructionFailed(item *models.TOptionAssetInstruction, cause error) error {
	item.RetryCount++
	item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED)
	item.NextRetryAt = time.Now().Add(optionAssetRetryDelay(item.RetryCount)).Unix()
	if item.RetryCount >= 20 {
		item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW)
		item.NextRetryAt = 0
	}
	item.LastErrorMsg = cause.Error()
	item.UpdateTimes = time.Now().Unix()
	return l.svcCtx.OptionAssetInstructionModel.Update(l.ctx, item)
}

func (l *ProcessAssetInstructionsLogic) completeFundingTransition(item *models.TOptionAssetInstruction) error {
	if item.OrderId == 0 ||
		item.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE) {
		return nil
	}
	order, err := l.svcCtx.OptionOrderModel.FindOne(l.ctx, item.OrderId)
	if err != nil {
		return err
	}
	if order.Status == int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		return applogic.MatchFundedOrder(l.ctx, l.svcCtx, order)
	}
	if order.Status != int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
		return nil
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, order.ContractId)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	admitToBook := contract.TenantId == order.TenantId &&
		contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
		contract.IsDeleted != int64(common.YesNo_YES_NO_YES) &&
		now >= contract.ListTime && (contract.ExpireTime == 0 || now < contract.ExpireTime)

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := orderModel.FindOne(ctx, order.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
			*order = *current
			return nil
		}
		if admitToBook {
			current.Status = int64(option.OrderStatus_ORDER_STATUS_PENDING)
		} else {
			current.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRING)
			current.CancelReason = "CONTRACT_NOT_TRADABLE_AFTER_FUNDING"
			current.CancelTime = now
			instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: current.TenantId, InstructionNo: current.OrderNo + "-FUNDING-RELEASE",
				BizNo: current.OrderNo, OrderId: current.Id, UserId: current.UserId, AccountId: current.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: current.OrderNo, Coin: current.FeeCoin, Amount: current.MarginAmount,
				StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		current.UpdateTimes = now
		if err := orderModel.Update(ctx, current); err != nil {
			return err
		}
		*order = *current
		return nil
	})
	if err != nil {
		return err
	}
	publishOptionOrderChanged(l.ctx, l.svcCtx, order)
	if order.Status == int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		return applogic.MatchFundedOrder(l.ctx, l.svcCtx, order)
	}
	return nil
}

func (l *ProcessAssetInstructionsLogic) completeOrderTransition(item *models.TOptionAssetInstruction) error {
	if item.OrderId == 0 ||
		item.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN) {
		return nil
	}
	order, err := l.svcCtx.OptionOrderModel.FindOne(l.ctx, item.OrderId)
	if err != nil {
		return err
	}
	switch option.OrderStatus(order.Status) {
	case option.OrderStatus_ORDER_STATUS_CANCELING:
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	case option.OrderStatus_ORDER_STATUS_EXPIRING:
		order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRED)
	default:
		return nil
	}
	order.MarginAmount = order.MarginAmount.Sub(item.Amount)
	if order.MarginAmount.IsNegative() {
		order.MarginAmount = decimal.Zero
	}
	order.UpdateTimes = time.Now().Unix()
	if err := l.svcCtx.OptionOrderModel.Update(l.ctx, order); err != nil {
		return err
	}
	publishOptionOrderChanged(l.ctx, l.svcCtx, order)
	return nil
}

func (l *ProcessAssetInstructionsLogic) completeSettlementTransition(item *models.TOptionAssetInstruction) error {
	if item.BizNo == "" || item.PositionId == 0 {
		return nil
	}
	instructions, err := l.svcCtx.OptionAssetInstructionModel.FindByBizNo(l.ctx, item.TenantId, item.BizNo)
	if err != nil {
		return err
	}
	if len(instructions) == 0 {
		return nil
	}
	successCount := int64(0)
	allSucceeded := true
	for _, instruction := range instructions {
		if instruction.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) {
			successCount++
		} else {
			allSucceeded = false
		}
	}
	batch, err := l.svcCtx.OptionSettlementBatchModel.FindOneByTenantIdBatchNo(l.ctx, item.TenantId, item.BizNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		return err
	}
	batch.SuccessCount = successCount
	batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_ASSET_PROCESSING)
	if allSucceeded {
		batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_RECONCILING)
	}
	batch.UpdateTimes = time.Now().Unix()
	if err := l.svcCtx.OptionSettlementBatchModel.Update(l.ctx, batch); err != nil {
		return err
	}
	if !allSucceeded {
		return nil
	}
	settlement, err := l.svcCtx.OptionSettlementModel.FindOneByTenantIdSettlementNo(l.ctx, item.TenantId, item.BizNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		return err
	}
	if settlement.Status == int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) {
		return nil
	}
	settlement.Status = int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE)
	settlement.UpdateTimes = time.Now().Unix()
	if err := l.svcCtx.OptionSettlementModel.Update(l.ctx, settlement); err != nil {
		return err
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, settlement.ContractId)
	if err != nil {
		return err
	}
	contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_SETTLED)
	contract.UpdateTimes = time.Now().Unix()
	if err := l.svcCtx.OptionContractModel.Update(l.ctx, contract); err != nil {
		return err
	}
	batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE)
	batch.UpdateTimes = time.Now().Unix()
	return l.svcCtx.OptionSettlementBatchModel.Update(l.ctx, batch)
}

func (l *ProcessAssetInstructionsLogic) execute(item *models.TOptionAssetInstruction) error {
	var baseCode int32
	var baseMsg string
	switch option.AssetInstructionAction(item.Action) {
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE:
		freezeBizNo := item.TargetBizNo
		if freezeBizNo == "" {
			freezeBizNo = item.InstructionNo
		}
		resp, err := l.svcCtx.AssetClient.FreezeAsset(l.ctx, &asset.FreezeAssetReq{
			TenantId: item.TenantId, UserId: item.UserId,
			WalletType: common.WalletType_WALLET_TYPE_OPTION,
			Coin:       item.Coin, Amount: item.Amount.String(),
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
			BizId: item.Id, BizNo: freezeBizNo, Remark: "option asset instruction freeze",
		})
		if err != nil {
			return err
		}
		if resp != nil && resp.Base != nil {
			baseCode, baseMsg = resp.Base.Code, resp.Base.Msg
		}
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN:
		resp, err := l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{
			TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_OPTION,
			TargetBizNo: item.TargetBizNo, Amount: item.Amount.String(),
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
			BizId: item.Id, BizNo: item.InstructionNo, Remark: "option asset instruction deduct frozen",
		})
		if err != nil {
			return err
		}
		if resp != nil && resp.Base != nil {
			baseCode, baseMsg = resp.Base.Code, resp.Base.Msg
		}
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN:
		resp, err := l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{
			TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_OPTION,
			TargetBizNo: item.TargetBizNo, Amount: item.Amount.String(),
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_CANCEL_ORDER,
			BizId: item.Id, BizNo: item.InstructionNo, Remark: "option asset instruction release frozen",
		})
		if err != nil {
			return err
		}
		if resp != nil && resp.Base != nil {
			baseCode, baseMsg = resp.Base.Code, resp.Base.Msg
		}
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE:
		resp, err := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
			TenantId: item.TenantId, UserId: item.UserId,
			WalletType: common.WalletType_WALLET_TYPE_OPTION,
			Coin:       item.Coin, Amount: item.Amount.String(),
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
			BizId: item.Id, BizNo: item.InstructionNo, Remark: "option asset instruction credit",
		})
		if err != nil {
			return err
		}
		if resp != nil && resp.Base != nil {
			baseCode, baseMsg = resp.Base.Code, resp.Base.Msg
		}
	case option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE:
		resp, err := l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{
			TenantId: item.TenantId, UserId: item.UserId,
			WalletType: common.WalletType_WALLET_TYPE_OPTION,
			Coin:       item.Coin, Amount: item.Amount.String(),
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
			BizId: item.Id, BizNo: item.InstructionNo, Remark: "option asset instruction debit",
		})
		if err != nil {
			return err
		}
		if resp != nil && resp.Base != nil {
			baseCode, baseMsg = resp.Base.Code, resp.Base.Msg
		}
	default:
		return fmt.Errorf("unsupported option asset instruction action: %d", item.Action)
	}
	if baseCode != 200 {
		return fmt.Errorf("asset rejected option instruction: code=%d msg=%s", baseCode, baseMsg)
	}
	return nil
}

func optionAssetRetryDelay(retryCount int64) time.Duration {
	if retryCount < 1 {
		retryCount = 1
	}
	if retryCount > 10 {
		retryCount = 10
	}
	return time.Duration(1<<uint(retryCount-1)) * time.Second
}
