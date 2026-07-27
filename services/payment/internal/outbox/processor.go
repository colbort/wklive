package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type rechargeCreditPayload struct {
	TenantID   int64  `json:"tenantId"`
	UserID     int64  `json:"userId"`
	WalletType int64  `json:"walletType"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
	Remark     string `json:"remark"`
}

func Start(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				process(ctx, svcCtx)
			}
		}
	}()
}

func process(ctx context.Context, svcCtx *svc.ServiceContext) {
	rows, err := svcCtx.PayOutboxModel.FindPending(ctx, time.Now().UnixMilli(), 100)
	if err != nil {
		logx.WithContext(ctx).Errorf("load payment outbox failed: %v", err)
		return
	}
	for _, row := range rows {
		if row.EventType != "PAYMENT_RECHARGE_CREDIT" {
			continue
		}
		var payload rechargeCreditPayload
		if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
			markFailed(ctx, svcCtx, row.Id, err.Error())
			continue
		}
		amount, err := decimal.NewFromString(payload.Amount)
		if err != nil || !amount.IsPositive() {
			markFailed(ctx, svcCtx, row.Id, "invalid recharge credit decimal amount")
			continue
		}
		resp, err := svcCtx.AssetCli.AddAvailable(ctx, &asset.AddAvailableReq{
			TenantId: payload.TenantID, UserId: payload.UserID,
			WalletType: common.WalletType(payload.WalletType), Coin: payload.Currency,
			Amount: amount.String(), BizType: asset.BizType_BIZ_TYPE_PAYMENT,
			SceneType: asset.SceneType_SCENE_TYPE_RECHARGE, BizId: row.AggregateId,
			BizNo: row.AggregateNo, Remark: payload.Remark,
		})
		if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			if err == nil {
				err = fmt.Errorf("asset credit response is unsuccessful")
			}
			markFailed(ctx, svcCtx, row.Id, err.Error())
			continue
		}
		row.Status = int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_SUCCESS)
		row.UpdateTimes = time.Now().UnixMilli()
		if err := svcCtx.PayOutboxModel.Update(ctx, row); err != nil {
			continue
		}
		if order, err := svcCtx.RechargeOrderModel.FindOne(ctx, row.AggregateId); err == nil {
			order.CreditStatus = int64(payment.CreditStatus_CREDIT_STATUS_SUCCESS)
			order.CreditedTime = row.UpdateTimes
			order.LastCreditError = ""
			order.UpdateTimes = row.UpdateTimes
			_ = svcCtx.RechargeOrderModel.Update(ctx, order)
		}
	}
}

func markFailed(ctx context.Context, svcCtx *svc.ServiceContext, id int64, message string) {
	row, err := svcCtx.PayOutboxModel.FindOne(ctx, id)
	if err != nil {
		return
	}
	row.Status = int64(payment.PayOutboxStatus_PAY_OUTBOX_STATUS_FAILED)
	row.RetryCount++
	delay := time.Duration(row.RetryCount) * time.Second
	if delay > time.Minute {
		delay = time.Minute
	}
	row.NextRetryAt = time.Now().Add(delay).UnixMilli()
	row.LastErrorMsg = message
	row.UpdateTimes = time.Now().UnixMilli()
	_ = svcCtx.PayOutboxModel.Update(ctx, row)
	if order, err := svcCtx.RechargeOrderModel.FindOne(ctx, row.AggregateId); err == nil {
		order.CreditStatus = int64(payment.CreditStatus_CREDIT_STATUS_FAILED)
		order.CreditRetryCount = row.RetryCount
		order.LastCreditError = message
		order.UpdateTimes = row.UpdateTimes
		_ = svcCtx.RechargeOrderModel.Update(ctx, order)
	}
}
