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
		if contract.ListTime != message.DueAt || now < contract.ListTime ||
			contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) {
			return nil
		}
		contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_TRADING)
		contract.UpdateTimes = now
		return svcCtx.OptionContractModel.Update(ctx, contract)
	case delayqueue.ActionExpireContract:
		if contract.ExpireTime != message.DueAt || now < contract.ExpireTime ||
			contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) {
			return nil
		}
		if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED) {
			contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED)
			contract.UpdateTimes = now
			if err := svcCtx.OptionContractModel.Update(ctx, contract); err != nil {
				return err
			}
		}
		if err := logic.expireContractOrders(contract, now); err != nil {
			return err
		}
		if contract.IsAutoExercise == 1 {
			if err := logic.autoExerciseContract(contract, now); err != nil {
				return err
			}
		}
		return logic.settleContract(contract, now)
	default:
		return nil
	}
}
