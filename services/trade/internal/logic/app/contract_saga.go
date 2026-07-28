package applogic

import (
	"context"
	"errors"
	"fmt"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type contractAssetStep struct {
	suffix string
	action trade.SettlementInstructionAction
	amount decimal.Decimal
	stepNo int64
}

func settlementInstructionLeaseOwned(current, claimed *models.TTradeSettlementInstruction) bool {
	return current != nil && claimed != nil &&
		current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING) &&
		current.UpdateTimes == claimed.UpdateTimes
}

func deliveryAssetSteps(margin, pnl, fee decimal.Decimal) []contractAssetStep {
	candidates := []contractAssetStep{
		{suffix: "MARGIN", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, amount: margin, stepNo: 1},
		{suffix: "LOSS", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, amount: decimalMaxZero(pnl.Neg()), stepNo: 2},
		{suffix: "FEE", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, amount: fee, stepNo: 2},
		{suffix: "PROFIT", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, amount: decimalMaxZero(pnl), stepNo: 3},
	}
	out := make([]contractAssetStep, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.amount.IsPositive() {
			out = append(out, candidate)
		}
	}
	return out
}

func legacyDeliveryAssetSteps(margin, pnl, fee decimal.Decimal) []contractAssetStep {
	candidates := []contractAssetStep{
		{suffix: "LOSS", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, amount: decimalMaxZero(pnl.Neg()), stepNo: 1},
		{suffix: "FEE", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS, amount: fee, stepNo: 1},
		{suffix: "MARGIN", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, amount: margin, stepNo: 2},
		{suffix: "PROFIT", action: trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE, amount: decimalMaxZero(pnl), stepNo: 2},
	}
	out := make([]contractAssetStep, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.amount.IsPositive() {
			out = append(out, candidate)
		}
	}
	return out
}

func matchesDeliveryAssetStep(item *models.TTradeSettlementInstruction, steps []contractAssetStep) bool {
	for _, step := range steps {
		if item.InstructionNo == item.BizId+"-"+step.suffix && item.Action == int64(step.action) && item.StepNo == step.stepNo && item.Amount.Equal(step.amount) {
			return true
		}
	}
	return false
}

func executeSimpleAssetInstruction(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TTradeSettlementInstruction, remark string) error {
	var resp *asset.ChangeAssetResp
	var err error
	switch trade.SettlementInstructionAction(item.Action) {
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE:
		resp, err = svcCtx.AssetClient.AddAvailable(ctx, &asset.AddAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: item.Asset, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: item.Id, BizNo: item.InstructionNo, Remark: remark})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS:
		resp, err = svcCtx.AssetClient.SubAvailable(ctx, &asset.SubAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: item.Asset, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH, BizId: item.Id, BizNo: item.InstructionNo, Remark: remark})
	default:
		return fmt.Errorf("unsupported contract saga action: %d", item.Action)
	}
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return errors.New("contract saga Asset returned empty response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("contract saga Asset rejected: code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return nil
}

func failContractSagaInstruction(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TTradeSettlementInstruction, cause error, updateBiz func(context.Context, sqlx.SqlConn, *models.TTradeSettlementInstruction, bool, int64) error) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(ctx, svcCtx.DB, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		im := models.NewTTradeSettlementInstructionModel(conn, svcCtx.Config.CacheRedis)
		current, err := im.FindOneForUpdate(txCtx, item.Id)
		if err != nil {
			return err
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return nil
		}
		current.RetryCount++
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)
		current.NextRetryAt = now + helpers.TradeEventRetryDelay(current.RetryCount).Milliseconds()
		manual := current.RetryCount >= 20
		if manual {
			current.Status, current.NextRetryAt = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW), 0
		}
		current.LastErrorMsg, current.UpdateTimes = cause.Error(), now
		if err = im.Update(txCtx, current); err != nil {
			return err
		}
		return updateBiz(txCtx, conn, current, manual, now)
	})
}
