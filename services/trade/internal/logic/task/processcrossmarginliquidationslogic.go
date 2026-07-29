package tasklogic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const crossAccountLiquidationBizType = "cross_liquidation"

type ProcessCrossMarginLiquidationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProcessCrossMarginLiquidationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessCrossMarginLiquidationsLogic {
	return &ProcessCrossMarginLiquidationsLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ProcessCrossMarginLiquidationsLogic) ProcessRiskSnapshots(tenantID int64) error {
	cursor := int64(0)
	for {
		snapshots, _, err := l.svcCtx.ContractMarginSnapshotModel.FindPage(l.ctx, tenantID, 0, "", cursor, 100)
		if err != nil {
			return err
		}
		if len(snapshots) == 0 {
			return nil
		}
		for _, snapshot := range snapshots {
			cursor = snapshot.Id
			if snapshot.PositionCount <= 0 || snapshot.RiskRate.LessThan(decimal.NewFromInt(1)) {
				continue
			}
			if err = l.processSnapshot(snapshot); err != nil {
				return err
			}
		}
		if len(snapshots) < 100 {
			return nil
		}
	}
}

func (l *ProcessCrossMarginLiquidationsLogic) Recover(tenantID, limit int64) error {
	rows, err := l.svcCtx.ContractAccountLiqModel.FindRecoverable(l.ctx, tenantID, limit)
	if err != nil {
		return err
	}
	var firstErr error
	for _, row := range rows {
		if err = l.process(row); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (l *ProcessCrossMarginLiquidationsLogic) processSnapshot(snapshot *models.TContractMarginSnapshot) error {
	if snapshot == nil || snapshot.PositionCount <= 0 || snapshot.RiskRate.LessThan(decimal.NewFromInt(1)) {
		return nil
	}
	batch, err := l.svcCtx.ContractAccountLiqModel.FindActiveByRiskUnit(l.ctx, snapshot.TenantId, snapshot.UserId, snapshot.MarginAsset)
	if errors.Is(err, models.ErrNotFound) {
		now := utils.NowMillis()
		batch = &models.TContractAccountLiquidation{
			TenantId: snapshot.TenantId, LiquidationNo: crossAccountLiquidationNo(snapshot),
			UserId: snapshot.UserId, MarginAsset: snapshot.MarginAsset,
			MarginSnapshotId: snapshot.Id, MarginSnapshotVersion: snapshot.Version,
			AssetVersion: snapshot.AssetVersion, WalletBalance: snapshot.WalletBalance,
			PositionMargin: snapshot.PositionMargin, MaintenanceMargin: snapshot.MaintenanceMargin,
			AccountEquity: snapshot.AccountEquity, RiskRate: snapshot.RiskRate,
			PositionCount: snapshot.PositionCount,
			Status:        models.ContractAccountLiquidationStatusPending,
			Reason:        "cross margin account maintenance margin breached",
			Version:       1, CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := l.svcCtx.ContractAccountLiqModel.Insert(l.ctx, batch)
		err = insertErr
		if err == nil {
			batch.Id, err = result.LastInsertId()
		}
		if err != nil {
			// A concurrent ProcessPositions worker can win the unique batch
			// insert. Resolve through the active risk-unit identity.
			batch, err = l.svcCtx.ContractAccountLiqModel.FindActiveByRiskUnit(l.ctx, snapshot.TenantId, snapshot.UserId, snapshot.MarginAsset)
		}
	}
	if err != nil {
		return err
	}
	return l.process(batch)
}

func crossAccountLiquidationNo(snapshot *models.TContractMarginSnapshot) string {
	return fmt.Sprintf("XLIQ-%d-%d", snapshot.Id, snapshot.Version)
}

func (l *ProcessCrossMarginLiquidationsLogic) process(batch *models.TContractAccountLiquidation) error {
	if batch == nil ||
		batch.Status == models.ContractAccountLiquidationStatusCompleted ||
		batch.Status == models.ContractAccountLiquidationStatusManualReview {
		return nil
	}
	// Closing the production gate prevents a new takeover. A Saga that already
	// locked positions or changed Asset must still be allowed to recover.
	if shouldHoldCrossAccountLiquidation(l.svcCtx.Config.AutomaticLiquidation.Enabled, batch.Status) {
		return l.markManual(batch.Id, "automatic account liquidation is disabled until P2 production acceptance")
	}
	if batch.Status == models.ContractAccountLiquidationStatusPending {
		if err := l.cancelRiskUnitOrders(batch); err != nil {
			return err
		}
		if err := l.prepareTakeover(batch); err != nil {
			return err
		}
		current, err := l.svcCtx.ContractAccountLiqModel.FindOne(l.ctx, batch.Id)
		if err != nil {
			return err
		}
		batch = current
	}
	if batch.Status == models.ContractAccountLiquidationStatusAssetSettling {
		if err := l.settleNetAssetChange(batch); err != nil {
			return err
		}
		current, err := l.svcCtx.ContractAccountLiqModel.FindOne(l.ctx, batch.Id)
		if err != nil {
			return err
		}
		batch = current
	}
	if batch.Status == models.ContractAccountLiquidationStatusInsuranceFund {
		if err := l.settleInsuranceFund(batch); err != nil {
			return err
		}
		current, err := l.svcCtx.ContractAccountLiqModel.FindOne(l.ctx, batch.Id)
		if err != nil {
			return err
		}
		batch = current
	}
	if batch.Status == models.ContractAccountLiquidationStatusADL {
		if err := l.settleADL(batch); err != nil {
			return err
		}
		current, err := l.svcCtx.ContractAccountLiqModel.FindOne(l.ctx, batch.Id)
		if err != nil {
			return err
		}
		batch = current
	}
	if batch.Status == models.ContractAccountLiquidationStatusClosing {
		feeSettled, err := l.settleLiquidationFee(batch)
		if err != nil {
			return err
		}
		if !feeSettled {
			return nil
		}
		return l.completeTakeover(batch)
	}
	return nil
}

func shouldHoldCrossAccountLiquidation(enabled bool, status int64) bool {
	return !enabled && status == models.ContractAccountLiquidationStatusPending
}

func (l *ProcessCrossMarginLiquidationsLogic) cancelRiskUnitOrders(batch *models.TContractAccountLiquidation) error {
	orderIDs, err := l.svcCtx.TradeOrderModel.FindCrossMarginCancelableOrderIDs(
		l.ctx, batch.TenantId, batch.UserId, batch.MarginAsset,
	)
	if err != nil {
		return err
	}
	for _, orderID := range orderIDs {
		order, err := beginSystemOrderTermination(l.ctx, l.svcCtx, orderID, "cross margin account liquidation", false)
		if err != nil {
			return err
		}
		if order == nil {
			continue
		}
		if err = unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, order, "cross margin account liquidation release"); err != nil {
			return err
		}
		if err = removeOrderBookOrder(l.svcCtx, l.ctx, order); err != nil {
			return err
		}
	}
	if err := NewProcessReservationReleasesLogic(l.ctx, l.svcCtx).Process(batch.TenantId); err != nil {
		return err
	}
	remaining, err := l.svcCtx.TradeAssetReservationModel.CountUnsettledCrossMarginByRiskUnit(
		l.ctx, batch.TenantId, batch.UserId, batch.MarginAsset,
	)
	if err != nil {
		return err
	}
	if remaining > 0 {
		return fmt.Errorf("cross margin risk unit still has %d unsettled order reservations", remaining)
	}
	return nil
}

func (l *ProcessCrossMarginLiquidationsLogic) prepareTakeover(batch *models.TContractAccountLiquidation) error {
	wallet, err := l.loadContractWallet(batch.TenantId, batch.UserId, batch.MarginAsset)
	if err != nil {
		return err
	}
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		bim := tx.ContractAccountLiquidationItem
		pm := tx.ContractPosition
		im := tx.TradeSettlementInstruction
		current, err := bm.FindOneForUpdate(ctx, batch.Id)
		if err != nil {
			return err
		}
		if current.Status != models.ContractAccountLiquidationStatusPending {
			return nil
		}
		positions, err := pm.FindCrossRiskUnitForUpdate(ctx, current.TenantId, current.UserId, current.MarginAsset)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return completeRecoveredCrossBatch(ctx, bm, current, "cross margin risk unit has no open position")
		}
		positionMargin, maintenance, pnl := decimal.Zero, decimal.Zero, decimal.Zero
		nominalFees := make([]decimal.Decimal, len(positions))
		for index, position := range positions {
			if position.MarkSnapshotId == "" || !position.MarkPrice.IsPositive() {
				return errors.New("cross account liquidation requires confirmed mark snapshots for every position")
			}
			contract, findErr := tx.TradeSymbolContract.
				FindOneByTenantIdSymbolId(ctx, position.TenantId, position.SymbolId)
			if findErr != nil {
				return findErr
			}
			values, calcErr := contractmath.CalculateTradeValues(position.ContractValueType, position.Qty, contract.ContractSize, position.MarkPrice)
			if calcErr != nil {
				return calcErr
			}
			nominalFees[index] = contractmath.RoundDebit(values.SettlementNotional.Mul(contract.LiquidationFeeRate))
			positionMargin = positionMargin.Add(position.PositionMargin)
			maintenance = maintenance.Add(position.MaintenanceMargin)
			pnl = pnl.Add(position.UnrealizedPnl)
		}
		equity, _, risk := calculateCrossAccountRisk(wallet.total, wallet.available, positionMargin, pnl, maintenance)
		if !maintenance.IsPositive() || risk.LessThan(decimal.NewFromInt(1)) {
			return completeRecoveredCrossBatch(ctx, bm, current, "cross margin account risk recovered before takeover")
		}
		gross := positionMargin.Add(pnl)
		nominalFee := decimal.Zero
		for _, fee := range nominalFees {
			nominalFee = nominalFee.Add(fee)
		}
		fee, credit, debit, deficit := calculateCrossTakeoverSettlement(gross, equity, nominalFee, wallet.total)
		fees := allocateCrossLiquidationFees(nominalFees, fee)
		now := utils.NowMillis()
		for index, position := range positions {
			position.Status = int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING)
			position.Version++
			position.UpdateTimes = now
			if err = pm.Update(ctx, position); err != nil {
				return err
			}
			if err = bim.InsertIdempotent(ctx, &models.TContractAccountLiquidationItem{
				TenantId: current.TenantId, AccountLiquidationId: current.Id,
				LiquidationNo: current.LiquidationNo, PositionId: position.Id,
				PositionVersion: position.Version, SymbolId: position.SymbolId,
				PositionSide: position.PositionSide, TriggerQty: position.Qty,
				TriggerMarkPrice: position.MarkPrice, TriggerSnapshotId: position.MarkSnapshotId,
				PositionMargin: position.PositionMargin, MaintenanceMargin: position.MaintenanceMargin,
				RealizedPnl: position.UnrealizedPnl, LiquidationFee: fees[index],
				Status:      models.ContractAccountLiquidationItemStatusLocked,
				CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		if credit.IsPositive() || debit.IsPositive() {
			action, amount := trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, credit
			if debit.IsPositive() {
				action, amount = trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, debit
			}
			if err = insertSettlementInstructionIdempotent(ctx, im, &models.TTradeSettlementInstruction{
				TenantId: current.TenantId, InstructionNo: current.LiquidationNo + "-NET",
				BizType: crossAccountLiquidationBizType, BizId: current.LiquidationNo,
				BatchNo: current.LiquidationNo, UserId: current.UserId,
				Action: int64(action), Asset: current.MarginAsset, Amount: amount,
				StepNo: 1, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING),
				NextRetryAt: now, CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		if fee.IsPositive() {
			if err = insertSettlementInstructionIdempotent(ctx, im, crossLiquidationFeeInstruction(current, fee, now)); err != nil {
				return err
			}
		}
		current.AssetVersion = wallet.version
		current.WalletBalance = wallet.total
		current.PositionMargin = positionMargin
		current.MaintenanceMargin = maintenance
		current.AccountEquity = equity
		current.RiskRate = risk
		current.GrossSettlement = gross
		current.LiquidationFee = fee
		current.UserCredit = credit
		current.UserDebit = debit
		current.DeficitAmount = deficit
		current.PositionCount = int64(len(positions))
		current.Status = models.ContractAccountLiquidationStatusAssetSettling
		if credit.IsZero() && debit.IsZero() {
			current.Status = crossAccountStageAfterNet(current)
		}
		current.Reason = "cross margin account positions taken over"
		if deficit.IsPositive() {
			current.Reason = "negative cross margin account positions taken over; awaiting insurance fund"
		}
		current.StartedAt = now
		current.Version++
		current.UpdateTimes = now
		return bm.Update(ctx, current)
	})
}

func calculateCrossTakeoverSettlement(gross, equity, nominalFee, walletTotal decimal.Decimal) (fee, credit, debit, deficit decimal.Decimal) {
	fee = decimalMin(decimalMaxZero(nominalFee), decimalMaxZero(equity))
	net := gross.Sub(fee)
	credit = decimalMaxZero(net)
	requiredDebit := decimalMaxZero(net.Neg())
	debit = decimalMin(requiredDebit, decimalMaxZero(walletTotal))
	deficit = decimalMaxZero(requiredDebit.Sub(debit))
	return fee, credit, debit, deficit
}

func crossAccountStageAfterNet(batch *models.TContractAccountLiquidation) int64 {
	if batch != nil && batch.DeficitAmount.IsPositive() {
		return models.ContractAccountLiquidationStatusInsuranceFund
	}
	return models.ContractAccountLiquidationStatusClosing
}

type crossContractWallet struct {
	total     decimal.Decimal
	available decimal.Decimal
	version   int64
}

func (l *ProcessCrossMarginLiquidationsLogic) loadContractWallet(tenantID, userID int64, marginAsset string) (*crossContractWallet, error) {
	resp, err := l.svcCtx.AssetAdminClient.GetUserAssetDetail(l.ctx, &asset.GetUserAssetDetailReq{
		TenantId: tenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: marginAsset,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return nil, errors.New("cross account liquidation Asset wallet query rejected")
	}
	total, err := decimal.NewFromString(resp.GetData().GetTotalAmount())
	if err != nil {
		return nil, err
	}
	available, err := decimal.NewFromString(resp.GetData().GetAvailableAmount())
	if err != nil {
		return nil, err
	}
	return &crossContractWallet{total: total, available: available, version: resp.GetData().GetVersion()}, nil
}

func completeRecoveredCrossBatch(ctx context.Context, model models.TContractAccountLiquidationModel, batch *models.TContractAccountLiquidation, reason string) error {
	now := utils.NowMillis()
	batch.Status = models.ContractAccountLiquidationStatusCompleted
	batch.Reason = reason
	batch.CompletedAt = now
	batch.Version++
	batch.UpdateTimes = now
	return model.Update(ctx, batch)
}

func allocateCrossLiquidationFees(nominal []decimal.Decimal, total decimal.Decimal) []decimal.Decimal {
	out := make([]decimal.Decimal, len(nominal))
	if len(nominal) == 0 || !total.IsPositive() {
		return out
	}
	nominalTotal := decimal.Zero
	for _, amount := range nominal {
		nominalTotal = nominalTotal.Add(decimalMaxZero(amount))
	}
	if !nominalTotal.IsPositive() {
		return out
	}
	remaining := total
	for index, amount := range nominal {
		if index == len(nominal)-1 {
			out[index] = remaining
			break
		}
		out[index] = total.Mul(decimalMaxZero(amount)).Div(nominalTotal)
		remaining = remaining.Sub(out[index])
	}
	return out
}

func (l *ProcessCrossMarginLiquidationsLogic) settleNetAssetChange(batch *models.TContractAccountLiquidation) error {
	instruction, err := l.svcCtx.TradeSettlementInstrModel.FindOneByTenantIdInstructionNo(l.ctx, batch.TenantId, batch.LiquidationNo+"-NET")
	if errors.Is(err, models.ErrNotFound) && batch.UserCredit.IsZero() && batch.UserDebit.IsZero() {
		return l.checkpointAfterNet(batch.Id)
	}
	if err != nil {
		return err
	}
	if instruction.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
		return l.checkpointAfterNet(batch.Id)
	}
	claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, instruction.Id, utils.NowMillis())
	if err != nil || !claimed {
		return err
	}
	instruction.UpdateTimes = lease
	if err = executeSimpleAssetInstruction(l.ctx, l.svcCtx, instruction, "cross margin account liquidation net settlement"); err != nil {
		failErr := failContractSagaInstruction(l.ctx, l.svcCtx, instruction, err, func(ctx context.Context, tx *models.TransactionModels, current *models.TTradeSettlementInstruction, manual bool, now int64) error {
			bm := tx.ContractAccountLiquidation
			locked, findErr := bm.FindOneForUpdate(ctx, batch.Id)
			if findErr != nil {
				return findErr
			}
			locked.Reason = current.LastErrorMsg
			if manual {
				locked.Status = models.ContractAccountLiquidationStatusManualReview
			}
			locked.Version++
			locked.UpdateTimes = now
			return bm.Update(ctx, locked)
		})
		if failErr != nil {
			return failErr
		}
		return err
	}
	return l.markNetAssetSucceeded(batch.Id, instruction)
}

func (l *ProcessCrossMarginLiquidationsLogic) markNetAssetSucceeded(batchID int64, claimed *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		im := tx.TradeSettlementInstruction
		bm := tx.ContractAccountLiquidation
		current, err := im.FindOneForUpdate(ctx, claimed.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			if !settlementInstructionLeaseOwned(current, claimed) {
				return errors.New("cross account settlement instruction lease was lost")
			}
			current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
			current.NextRetryAt = 0
			current.LastErrorMsg = ""
			current.UpdateTimes = now
			if err = im.Update(ctx, current); err != nil {
				return err
			}
		}
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status == models.ContractAccountLiquidationStatusAssetSettling {
			batch.Status = crossAccountStageAfterNet(batch)
			batch.Reason = "cross margin account Asset net settlement completed"
			batch.Version++
			batch.UpdateTimes = now
			return bm.Update(ctx, batch)
		}
		return nil
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) checkpointAfterNet(batchID int64) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status != models.ContractAccountLiquidationStatusAssetSettling {
			return nil
		}
		batch.Status = crossAccountStageAfterNet(batch)
		batch.Version++
		batch.UpdateTimes = now
		return bm.Update(ctx, batch)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) settleInsuranceFund(batch *models.TContractAccountLiquidation) error {
	if batch == nil || batch.Status != models.ContractAccountLiquidationStatusInsuranceFund {
		return nil
	}
	if !batch.DeficitAmount.IsPositive() {
		return l.checkpointCrossAccountStage(batch.Id, models.ContractAccountLiquidationStatusClosing, "cross account has no insurance deficit")
	}
	fund, err := l.svcCtx.ContractInsuranceFundModel.FindEnabled(l.ctx, batch.TenantId, 0, batch.MarginAsset)
	if errors.Is(err, models.ErrNotFound) {
		return l.markManual(batch.Id, "no enabled asset-level insurance fund account for cross margin deficit")
	}
	if err != nil {
		return err
	}
	resp, err := l.svcCtx.AssetClient.CoverInsuranceDeficit(l.ctx, &asset.CoverInsuranceDeficitReq{
		TenantId:        batch.TenantId,
		Coin:            batch.MarginAsset,
		RequestedAmount: batch.DeficitAmount.String(),
		LiquidationId:   batch.Id,
		LiquidationNo:   batch.LiquidationNo + "-INSURANCE",
		Remark:          "cross margin account liquidation insurance fund cover",
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return errors.New("cross account insurance fund returned an empty response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("cross account insurance fund rejected: %s", resp.GetBase().GetMsg())
	}
	covered, err := decimal.NewFromString(resp.GetCoveredAmount())
	if err != nil {
		return err
	}
	remaining, err := decimal.NewFromString(resp.GetRemainingAmount())
	if err != nil {
		return err
	}
	if covered.IsNegative() || remaining.IsNegative() ||
		!covered.Add(remaining).Equal(batch.DeficitAmount) {
		return errors.New("invalid cross account insurance fund coverage response")
	}
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		current, lockErr := bm.FindOneForUpdate(ctx, batch.Id)
		if lockErr != nil {
			return lockErr
		}
		if current.Status != models.ContractAccountLiquidationStatusInsuranceFund {
			return nil
		}
		current.InsuranceFundAmount = covered
		switch {
		case remaining.IsZero():
			current.Status = models.ContractAccountLiquidationStatusClosing
			current.Reason = "cross account deficit fully covered by insurance fund"
		case fund.AdlEnabled == 1:
			current.Status = models.ContractAccountLiquidationStatusADL
			current.Reason = "cross account insurance fund partially covered deficit; awaiting ADL"
		default:
			current.Status = models.ContractAccountLiquidationStatusManualReview
			current.Reason = "cross account insurance fund exhausted and ADL is disabled"
		}
		current.Version++
		current.UpdateTimes = now
		return bm.Update(ctx, current)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) settleADL(batch *models.TContractAccountLiquidation) error {
	if batch == nil || batch.Status != models.ContractAccountLiquidationStatusADL {
		return nil
	}
	required := decimalMaxZero(batch.DeficitAmount.Sub(batch.InsuranceFundAmount))
	if !required.IsPositive() {
		return l.checkpointCrossAccountStage(batch.Id, models.ContractAccountLiquidationStatusClosing, "cross account deficit covered before ADL")
	}
	items, err := l.svcCtx.ContractAccountLiqItemModel.FindByLiquidation(l.ctx, batch.TenantId, batch.Id, false)
	if err != nil {
		return err
	}
	targets, err := allocateCrossADLDeficits(items, required)
	if err != nil {
		return l.markManual(batch.Id, err.Error())
	}
	adlLogic := NewProcessLiquidationsLogic(l.ctx, l.svcCtx)
	totalRelief, totalQty := decimal.Zero, decimal.Zero
	for index, item := range items {
		target := targets[index]
		if !target.IsPositive() {
			continue
		}
		position, findErr := l.svcCtx.ContractPositionModel.FindOne(l.ctx, item.PositionId)
		if findErr != nil {
			return findErr
		}
		if position.Status != int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING) ||
			position.Version != item.PositionVersion || !position.Qty.Equal(item.TriggerQty) {
			return l.markManual(batch.Id, fmt.Sprintf("cross account ADL source position %d changed after takeover", item.PositionId))
		}
		contract, findErr := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, batch.TenantId, item.SymbolId)
		if findErr != nil {
			return findErr
		}
		bankruptcyPrice := item.BankruptcyPrice
		if item.DeficitAmount.IsPositive() && !item.DeficitAmount.Equal(target) {
			return l.markManual(batch.Id, fmt.Sprintf("cross account ADL target changed for item %d", item.Id))
		}
		if !bankruptcyPrice.IsPositive() {
			bankruptcyPrice, findErr = crossAccountBankruptcyPrice(position, contract.ContractSize, target)
			if findErr != nil {
				return l.markManual(batch.Id, findErr.Error())
			}
			if findErr = l.checkpointCrossADLItem(batch.Id, item.Id, target, bankruptcyPrice, item.AdlReliefAmount, item.AdlQty); findErr != nil {
				return findErr
			}
		}
		existingExecutions, findErr := l.svcCtx.ContractAdlExecutionModel.FindByLiquidation(
			l.ctx, batch.TenantId, -item.Id,
		)
		if findErr != nil {
			return findErr
		}
		for _, execution := range existingExecutions {
			if execution.Status == 5 {
				return l.markManual(
					batch.Id,
					fmt.Sprintf("cross account ADL execution %s requires manual review", execution.ExecutionNo),
				)
			}
		}
		bankrupt := cloneContractPosition(position)
		bankrupt.BankruptcyPrice = bankruptcyPrice
		synthetic := &models.TContractLiquidation{
			Id: -item.Id, LiquidationNo: fmt.Sprintf("%s-ITEM-%d", batch.LiquidationNo, item.Id),
			AdlQty: item.AdlQty,
		}
		adlQty, remaining, executeErr := adlLogic.executeADL(bankrupt, contract, synthetic, target)
		if executeErr != nil {
			return executeErr
		}
		relieved := decimalMaxZero(target.Sub(remaining))
		if err = l.checkpointCrossADLItem(batch.Id, item.Id, target, bankruptcyPrice, relieved, adlQty); err != nil {
			return err
		}
		totalRelief = totalRelief.Add(relieved)
		totalQty = totalQty.Add(adlQty)
		if !crossAmountCovered(target, relieved) {
			return l.markManual(batch.Id, fmt.Sprintf("insufficient ADL capacity for cross account item %d: remaining=%s", item.Id, decimalMaxZero(target.Sub(relieved)).String()))
		}
	}
	if !crossAmountCovered(required, totalRelief) {
		return l.markManual(batch.Id, fmt.Sprintf("insufficient ADL capacity for cross account: remaining=%s", decimalMaxZero(required.Sub(totalRelief)).String()))
	}
	return l.finishCrossADL(batch.Id, totalRelief, totalQty)
}

func allocateCrossADLDeficits(items []*models.TContractAccountLiquidationItem, total decimal.Decimal) ([]decimal.Decimal, error) {
	targets := make([]decimal.Decimal, len(items))
	if !total.IsPositive() {
		return targets, nil
	}
	totalLoss := decimal.Zero
	for _, item := range items {
		if item != nil && item.RealizedPnl.IsNegative() {
			totalLoss = totalLoss.Add(item.RealizedPnl.Neg())
		}
	}
	if !totalLoss.IsPositive() || total.GreaterThan(totalLoss) {
		return nil, fmt.Errorf("cross account deficit %s exceeds frozen losing-position PnL %s", total.String(), totalLoss.String())
	}
	remaining := total
	lastLosing := -1
	for index, item := range items {
		if item != nil && item.RealizedPnl.IsNegative() {
			lastLosing = index
		}
	}
	for index, item := range items {
		if item == nil || !item.RealizedPnl.IsNegative() {
			continue
		}
		if index == lastLosing {
			targets[index] = remaining
			break
		}
		targets[index] = total.Mul(item.RealizedPnl.Neg()).Div(totalLoss)
		remaining = remaining.Sub(targets[index])
	}
	return targets, nil
}

func crossAccountBankruptcyPrice(position *models.TContractPosition, contractSize, relief decimal.Decimal) (decimal.Decimal, error) {
	if position == nil || !position.Qty.IsPositive() || !position.OpenAvgPrice.IsPositive() ||
		!position.MarkPrice.IsPositive() || !contractSize.IsPositive() || !relief.IsPositive() {
		return decimal.Zero, errors.New("cross account ADL bankruptcy price requires positive frozen position facts")
	}
	markPnl := contractRealizedPnl(
		position.PositionSide, position.OpenAvgPrice, position.MarkPrice,
		position.Qty, contractSize, position.ContractValueType,
	)
	if !markPnl.IsNegative() || relief.GreaterThan(markPnl.Neg()) {
		return decimal.Zero, fmt.Errorf("cross account ADL relief %s exceeds position loss %s", relief.String(), markPnl.Neg().String())
	}
	targetPnl := markPnl.Add(relief)
	contracts := position.Qty.Mul(contractSize)
	var price decimal.Decimal
	if position.ContractValueType == int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE) {
		inverseOpen := decimal.NewFromInt(1).Div(position.OpenAvgPrice)
		var inversePrice decimal.Decimal
		if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_SHORT) {
			inversePrice = inverseOpen.Add(targetPnl.Div(contracts))
		} else {
			inversePrice = inverseOpen.Sub(targetPnl.Div(contracts))
		}
		if !inversePrice.IsPositive() {
			return decimal.Zero, errors.New("cross account ADL inverse bankruptcy boundary is not positive")
		}
		price = decimal.NewFromInt(1).Div(inversePrice)
	} else {
		delta := targetPnl.Div(contracts)
		if position.PositionSide == int64(trade.PositionSide_POSITION_SIDE_SHORT) {
			price = position.OpenAvgPrice.Sub(delta)
		} else {
			price = position.OpenAvgPrice.Add(delta)
		}
	}
	price = price.Round(18)
	if !price.IsPositive() {
		return decimal.Zero, errors.New("cross account ADL bankruptcy price is not positive")
	}
	return price, nil
}

func crossAmountCovered(required, actual decimal.Decimal) bool {
	if actual.GreaterThanOrEqual(required) {
		return true
	}
	return required.Sub(actual).LessThanOrEqual(decimal.New(1, -12))
}

func crossAmountsEqual(left, right decimal.Decimal) bool {
	return left.Sub(right).Abs().LessThanOrEqual(decimal.New(1, -12))
}

func (l *ProcessCrossMarginLiquidationsLogic) checkpointCrossADLItem(
	batchID, itemID int64,
	target, bankruptcyPrice, relief, qty decimal.Decimal,
) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		bim := tx.ContractAccountLiquidationItem
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status != models.ContractAccountLiquidationStatusADL {
			return nil
		}
		items, err := bim.FindByLiquidation(ctx, batch.TenantId, batch.Id, true)
		if err != nil {
			return err
		}
		found := false
		totalRelief, totalQty := decimal.Zero, decimal.Zero
		for _, item := range items {
			if item.Id == itemID {
				if item.DeficitAmount.IsPositive() && !item.DeficitAmount.Equal(target) {
					return errors.New("cross account ADL item target is immutable")
				}
				item.DeficitAmount = target
				item.BankruptcyPrice = bankruptcyPrice
				if relief.GreaterThan(item.AdlReliefAmount) {
					item.AdlReliefAmount = relief
				}
				if qty.GreaterThan(item.AdlQty) {
					item.AdlQty = qty
				}
				item.UpdateTimes = now
				if err = bim.UpdateADL(ctx, item); err != nil {
					return err
				}
				found = true
			}
			totalRelief = totalRelief.Add(item.AdlReliefAmount)
			totalQty = totalQty.Add(item.AdlQty)
		}
		if !found {
			return fmt.Errorf("cross account ADL item %d not found", itemID)
		}
		batch.AdlReliefAmount = totalRelief
		batch.AdlQty = totalQty
		batch.Reason = "cross account ADL progress checkpointed"
		batch.Version++
		batch.UpdateTimes = now
		return bm.Update(ctx, batch)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) finishCrossADL(batchID int64, relief, qty decimal.Decimal) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status != models.ContractAccountLiquidationStatusADL {
			return nil
		}
		required := decimalMaxZero(batch.DeficitAmount.Sub(batch.InsuranceFundAmount))
		if !crossAmountCovered(required, relief) {
			return errors.New("cross account ADL cannot close before the deficit is covered")
		}
		batch.AdlReliefAmount = relief
		batch.AdlQty = qty
		batch.Status = models.ContractAccountLiquidationStatusClosing
		batch.Reason = "cross account deficit covered by insurance fund and ADL"
		batch.Version++
		batch.UpdateTimes = now
		return bm.Update(ctx, batch)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) checkpointCrossAccountStage(batchID, status int64, reason string) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status == models.ContractAccountLiquidationStatusCompleted ||
			batch.Status == models.ContractAccountLiquidationStatusManualReview {
			return nil
		}
		batch.Status = status
		batch.Reason = reason
		batch.Version++
		batch.UpdateTimes = now
		return bm.Update(ctx, batch)
	})
}

func crossLiquidationFeeInstruction(batch *models.TContractAccountLiquidation, fee decimal.Decimal, now int64) *models.TTradeSettlementInstruction {
	return &models.TTradeSettlementInstruction{
		TenantId: batch.TenantId, InstructionNo: batch.LiquidationNo + "-FEE",
		BizType: crossAccountLiquidationBizType, BizId: batch.LiquidationNo,
		BatchNo: batch.LiquidationNo, UserId: batch.UserId,
		Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE),
		Asset:  batch.MarginAsset, Amount: fee, StepNo: 2,
		Status:      int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING),
		NextRetryAt: now, CreateTimes: now, UpdateTimes: now,
	}
}

func (l *ProcessCrossMarginLiquidationsLogic) settleLiquidationFee(batch *models.TContractAccountLiquidation) (bool, error) {
	if !batch.LiquidationFee.IsPositive() {
		return true, nil
	}
	instruction, err := l.svcCtx.TradeSettlementInstrModel.FindOneByTenantIdInstructionNo(l.ctx, batch.TenantId, batch.LiquidationNo+"-FEE")
	if errors.Is(err, models.ErrNotFound) {
		now := utils.NowMillis()
		err = l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
			return insertSettlementInstructionIdempotent(
				ctx,
				tx.TradeSettlementInstruction,
				crossLiquidationFeeInstruction(batch, batch.LiquidationFee, now),
			)
		})
		if err != nil {
			return false, err
		}
		instruction, err = l.svcCtx.TradeSettlementInstrModel.FindOneByTenantIdInstructionNo(l.ctx, batch.TenantId, batch.LiquidationNo+"-FEE")
	}
	if err != nil {
		return false, err
	}
	switch trade.SettlementInstructionStatus(instruction.Status) {
	case trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS:
		return true, nil
	case trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW:
		if err = l.markManual(batch.Id, "cross margin liquidation fee settlement requires manual review"); err != nil {
			return false, err
		}
		return false, nil
	}
	now := utils.NowMillis()
	if instruction.NextRetryAt > now {
		return false, nil
	}
	claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, instruction.Id, now)
	if err != nil || !claimed {
		return false, err
	}
	resp, err := l.svcCtx.AssetClient.CreditPlatformRevenue(l.ctx, &asset.CreditPlatformRevenueReq{
		TenantId: batch.TenantId, Coin: batch.MarginAsset,
		Amount:  batch.LiquidationFee.String(),
		BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_FEE,
		BizId: batch.Id, BizNo: batch.LiquidationNo + "-FEE",
		Remark: "cross margin account liquidation fee revenue",
	})
	if err != nil {
		return false, l.failLiquidationFee(batch, instruction, lease, err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return false, l.failLiquidationFee(batch, instruction, lease, errors.New("cross margin liquidation fee revenue rejected"))
	}
	instruction.UpdateTimes = lease
	if err = l.markLiquidationFeeSucceeded(instruction); err != nil {
		return false, err
	}
	return true, nil
}

func (l *ProcessCrossMarginLiquidationsLogic) failLiquidationFee(batch *models.TContractAccountLiquidation, instruction *models.TTradeSettlementInstruction, lease int64, cause error) error {
	instruction.UpdateTimes = lease
	return failContractSagaInstruction(l.ctx, l.svcCtx, instruction, cause, func(ctx context.Context, tx *models.TransactionModels, current *models.TTradeSettlementInstruction, manual bool, now int64) error {
		bm := tx.ContractAccountLiquidation
		locked, err := bm.FindOneForUpdate(ctx, batch.Id)
		if err != nil {
			return err
		}
		locked.Reason = current.LastErrorMsg
		if manual {
			locked.Status = models.ContractAccountLiquidationStatusManualReview
		}
		locked.Version++
		locked.UpdateTimes = now
		return bm.Update(ctx, locked)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) markLiquidationFeeSucceeded(claimed *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		im := tx.TradeSettlementInstruction
		current, err := im.FindOneForUpdate(ctx, claimed.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if !settlementInstructionLeaseOwned(current, claimed) {
			return errors.New("cross account fee instruction lease was lost")
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.NextRetryAt = 0
		current.LastErrorMsg = ""
		// Platform revenue is audited in t_asset_platform_flow. Mark this
		// instruction reconciled so the user Asset-flow scanner does not search
		// for a t_asset_flow row that intentionally does not exist.
		current.AssetFlowNo = "PLATFORM:" + current.InstructionNo
		current.ReconciledAt = now
		current.UpdateTimes = now
		return im.Update(ctx, current)
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) completeTakeover(batch *models.TContractAccountLiquidation) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		bim := tx.ContractAccountLiquidationItem
		pm := tx.ContractPosition
		hm := tx.ContractPositionHistory
		em := tx.BizTradeEvent
		current, err := bm.FindOneForUpdate(ctx, batch.Id)
		if err != nil {
			return err
		}
		if current.Status == models.ContractAccountLiquidationStatusCompleted {
			return nil
		}
		if current.Status != models.ContractAccountLiquidationStatusClosing {
			return errors.New("cross account liquidation is not ready to close")
		}
		items, err := bim.FindByLiquidation(ctx, current.TenantId, current.Id, true)
		if err != nil {
			return err
		}
		if int64(len(items)) != current.PositionCount {
			return errors.New("cross account liquidation item count mismatch")
		}
		for _, item := range items {
			if item.Status == models.ContractAccountLiquidationItemStatusClosed {
				continue
			}
			position, findErr := pm.FindOneForUpdate(ctx, item.PositionId)
			if findErr != nil {
				return findErr
			}
			if position.Status != int64(trade.PositionStatus_POSITION_STATUS_LIQUIDATING) ||
				position.Version != item.PositionVersion ||
				!position.Qty.Equal(item.TriggerQty) {
				return fmt.Errorf("cross account liquidation position %d changed after takeover", position.Id)
			}
			before := cloneContractPosition(position)
			position.Qty, position.AvailQty, position.FrozenQty = decimal.Zero, decimal.Zero, decimal.Zero
			position.PositionMargin, position.IsolatedMargin = decimal.Zero, decimal.Zero
			position.MaintenanceMargin, position.UnrealizedPnl = decimal.Zero, decimal.Zero
			position.LiquidationPrice, position.BankruptcyPrice = decimal.Zero, decimal.Zero
			position.RiskRate = decimal.Zero
			position.AdlRank = 0
			position.Status = int64(trade.PositionStatus_POSITION_STATUS_CLOSED)
			position.ClosedAt = now
			position.Version++
			position.UpdateTimes = now
			if err = pm.Update(ctx, position); err != nil {
				return err
			}
			if err = writeSystemPositionHistory(
				ctx, hm, before, position, current.CreateTimes,
				fmt.Sprintf("%s-%d", current.LiquidationNo, position.Id),
				trade.PositionActionType_POSITION_ACTION_TYPE_LIQUIDATION,
				item.RealizedPnl, item.LiquidationFee, item.TriggerMarkPrice,
				"cross margin account liquidation",
			); err != nil {
				return err
			}
			if err = bim.UpdateStatus(ctx, item.Id, models.ContractAccountLiquidationItemStatusClosed, now); err != nil {
				return err
			}
		}
		current.Status = models.ContractAccountLiquidationStatusCompleted
		current.Reason = "cross margin account liquidation completed"
		current.CompletedAt = now
		current.Version++
		current.UpdateTimes = now
		if err = bm.Update(ctx, current); err != nil {
			return err
		}
		_, err = em.Insert(ctx, &models.TBizTradeEvent{
			TenantId: current.TenantId, EventNo: current.LiquidationNo + "-COMPLETED",
			EventType: "CROSS_ACCOUNT_LIQUIDATION_COMPLETED",
			BizId:     current.LiquidationNo, BizType: crossAccountLiquidationBizType,
			UserId: current.UserId, ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE),
			Source:        int64(trade.SourceType_SOURCE_TYPE_TASK),
			EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
			MaxRetryCount: 20, NextRetryAt: now, Payload: "{}",
			CreateTimes: now, UpdateTimes: now,
		})
		return err
	})
}

func (l *ProcessCrossMarginLiquidationsLogic) markManual(batchID int64, reason string) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		bm := tx.ContractAccountLiquidation
		batch, err := bm.FindOneForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status == models.ContractAccountLiquidationStatusCompleted ||
			batch.Status == models.ContractAccountLiquidationStatusManualReview {
			return nil
		}
		batch.Status = models.ContractAccountLiquidationStatusManualReview
		batch.Reason = reason
		batch.Version++
		batch.UpdateTimes = now
		return bm.Update(ctx, batch)
	})
}
