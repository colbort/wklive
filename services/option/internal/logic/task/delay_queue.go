package tasklogic

import (
	"context"
	"errors"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/delayqueue"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartDelayQueue(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DelayQueue == nil {
		return
	}
	go svcCtx.DelayQueue.Consume(func(message delayqueue.Message) {
		if err := handleDelayMessage(ctx, svcCtx, message); err != nil {
			logx.WithContext(ctx).Errorf("option delay message failed, action=%s contractId=%d err=%v",
				message.Action, message.ContractID, err)
		}
	})
}

func handleDelayMessage(ctx context.Context, svcCtx *svc.ServiceContext, message delayqueue.Message) error {
	contract, err := svcCtx.OptionContractModel.FindOne(ctx, message.ContractID)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if contract.TenantId != message.TenantID {
		return nil
	}
	now := time.Now().Unix()
	logic := NewProcessContractLifecycleLogic(ctx, svcCtx)
	switch message.Action {
	case delayqueue.ActionListContract:
		_, err = logic.listContractIfEligible(
			message.ContractID, message.TenantID, message.DueAt, now,
		)
		return err
	case delayqueue.ActionCloseContractTrading:
		if contract.LastTradeTime != message.DueAt || now < contract.LastTradeTime {
			return nil
		}
		return logic.closeContractTrading(
			message.ContractID, message.TenantID, message.DueAt, now,
		)
	case delayqueue.ActionExpireContract:
		if contract.ExpireTime != message.DueAt || now < contract.ExpireTime ||
			contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) {
			return nil
		}
		if err := logic.closeContractTrading(
			message.ContractID, message.TenantID, contract.LastTradeTime, now,
		); err != nil {
			return err
		}
		if err := logic.expirePausedContracts(now); err != nil {
			return err
		}
		contract, err = svcCtx.OptionContractModel.FindOne(ctx, message.ContractID)
		if err != nil {
			return err
		}
		if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED) {
			return nil
		}
		if err := logic.expireContractOrders(contract, now); err != nil {
			return err
		}
		pendingExercise, err := svcCtx.OptionExerciseModel.HasPendingByContract(
			ctx, contract.TenantId, contract.Id,
		)
		if err != nil || pendingExercise {
			return err
		}
		settlementPrice, err := logic.lockSettlementPrice(contract, now)
		if err != nil || settlementPrice == nil ||
			settlementPrice.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED) {
			return err
		}
		if contract.IsAutoExercise == 1 {
			if err := logic.autoExerciseContract(contract, settlementPrice.DeliveryPrice, now); err != nil {
				return err
			}
		}
		if contract.DeliverTime > now {
			return nil
		}
		return logic.settleContract(contract, settlementPrice, now)
	default:
		return nil
	}
}
