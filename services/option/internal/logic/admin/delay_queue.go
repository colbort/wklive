package adminlogic

import (
	"time"

	"wklive/services/option/internal/delayqueue"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

func enqueueContractSchedules(svcCtx *svc.ServiceContext, contract *models.TOptionContract) []error {
	if svcCtx == nil || svcCtx.DelayQueue == nil || contract == nil {
		return nil
	}
	var errs []error
	if contract.ListTime > time.Now().Unix() {
		if err := svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionListContract, TenantID: contract.TenantId,
			ContractID: contract.Id, DueAt: contract.ListTime,
		}, time.Unix(contract.ListTime, 0)); err != nil {
			errs = append(errs, err)
		}
	}
	if contract.LastTradeTime > time.Now().Unix() {
		if err := svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionCloseContractTrading, TenantID: contract.TenantId,
			ContractID: contract.Id, DueAt: contract.LastTradeTime,
		}, time.Unix(contract.LastTradeTime, 0)); err != nil {
			errs = append(errs, err)
		}
	}
	if contract.ExpireTime > time.Now().Unix() {
		if err := svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionExpireContract, TenantID: contract.TenantId,
			ContractID: contract.Id, DueAt: contract.ExpireTime,
		}, time.Unix(contract.ExpireTime, 0)); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
