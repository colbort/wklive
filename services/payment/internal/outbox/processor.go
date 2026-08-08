package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/common/worklease"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

const (
	outboxBatchSize    = int64(100)
	outboxClaimLease   = 30 * time.Second
	assetCallTimeout   = 10 * time.Second
	maximumRetryDelay  = time.Minute
	processorTickDelay = time.Second
)

var errOutboxClaimLost = errors.New("payment outbox claim is no longer owned by this worker")

type rechargeCreditPayload struct {
	TenantID   int64  `json:"tenantId"`
	UserID     int64  `json:"userId"`
	WalletType int64  `json:"walletType"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
	Remark     string `json:"remark"`
}

func Start(ctx context.Context, svcCtx *svc.ServiceContext) {
	owner := worklease.NewOwner("payment")
	go func() {
		ticker := time.NewTicker(processorTickDelay)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				process(ctx, svcCtx, owner)
			}
		}
	}()
}

func process(ctx context.Context, svcCtx *svc.ServiceContext, owner string) {
	now := time.Now()
	rows, err := svcCtx.PayOutboxModel.ClaimPending(
		ctx,
		owner,
		now.UnixMilli(),
		now.Add(-outboxClaimLease).UnixMilli(),
		outboxBatchSize,
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("claim payment outbox failed: owner=%s err=%v", owner, err)
	}
	for _, row := range rows {
		processClaimed(ctx, svcCtx, owner, row)
	}
}

func processClaimed(ctx context.Context, svcCtx *svc.ServiceContext, owner string, row *models.TPayOutbox) {
	if row.EventType != "PAYMENT_RECHARGE_CREDIT" {
		persistFailure(ctx, svcCtx, owner, row, fmt.Errorf("unsupported payment outbox event type %q", row.EventType))
		return
	}

	var payload rechargeCreditPayload
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		persistFailure(ctx, svcCtx, owner, row, fmt.Errorf("decode recharge credit payload: %w", err))
		return
	}
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil || !amount.IsPositive() {
		persistFailure(ctx, svcCtx, owner, row, errors.New("invalid recharge credit decimal amount"))
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, assetCallTimeout)
	resp, callErr := svcCtx.AssetCli.AddAvailable(callCtx, &asset.AddAvailableReq{
		TenantId: payload.TenantID, UserId: payload.UserID,
		WalletType: common.WalletType(payload.WalletType), Coin: payload.Currency,
		Amount: amount.String(), BizType: asset.BizType_BIZ_TYPE_PAYMENT,
		SceneType: asset.SceneType_SCENE_TYPE_RECHARGE, BizId: row.AggregateId,
		BizNo: row.AggregateNo, Remark: payload.Remark,
	})
	cancel()
	if callErr != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		if callErr == nil {
			callErr = errors.New("asset credit response is unsuccessful")
		}
		persistFailure(ctx, svcCtx, owner, row, callErr)
		return
	}

	if err := finalizeSuccess(ctx, svcCtx, owner, row, payload); err != nil {
		if errors.Is(err, errOutboxClaimLost) {
			logx.WithContext(ctx).Infof("payment outbox success ignored after claim loss: id=%d owner=%s", row.Id, owner)
			return
		}
		// Asset is already credited. Releasing this claim as retryable is safe
		// because AddAvailable is idempotent for the stable payment BizNo.
		logx.WithContext(ctx).Errorf("finalize payment outbox success failed: id=%d owner=%s err=%v", row.Id, owner, err)
		persistFailure(ctx, svcCtx, owner, row, fmt.Errorf("finalize credited recharge: %w", err))
	}
}

func finalizeSuccess(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	owner string,
	row *models.TPayOutbox,
	payload rechargeCreditPayload,
) error {
	now := time.Now().UnixMilli()
	return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		outboxModel := models.NewTPayOutboxModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTRechargeOrderModel(conn, svcCtx.Config.CacheRedis)

		updated, err := outboxModel.MarkSuccess(txCtx, row, owner, now)
		if err != nil {
			return err
		}
		if !updated {
			return errOutboxClaimLost
		}
		identity, err := orderModel.FindCreditIdentityForUpdate(txCtx, row.AggregateId)
		if err != nil {
			return err
		}
		if identity.TenantId != payload.TenantID || identity.OrderNo != row.AggregateNo {
			return fmt.Errorf(
				"recharge order identity mismatch: order_id=%d tenant_id=%d order_no=%s",
				row.AggregateId,
				payload.TenantID,
				row.AggregateNo,
			)
		}
		updated, err = orderModel.MarkCreditSuccess(txCtx, row.AggregateId, identity, now)
		if err != nil {
			return err
		}
		if !updated {
			return errors.New("recharge order was not updated")
		}
		return nil
	})
}

func persistFailure(ctx context.Context, svcCtx *svc.ServiceContext, owner string, row *models.TPayOutbox, cause error) {
	if row == nil || cause == nil {
		return
	}
	retryCount := row.RetryCount + 1
	delay := retryDelay(retryCount)
	now := time.Now()
	nextRetryAt := now.Add(delay).UnixMilli()
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		outboxModel := models.NewTPayOutboxModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTRechargeOrderModel(conn, svcCtx.Config.CacheRedis)

		updated, err := outboxModel.MarkFailed(
			txCtx,
			row,
			owner,
			retryCount,
			nextRetryAt,
			now.UnixMilli(),
			cause.Error(),
		)
		if err != nil {
			return err
		}
		if !updated {
			return errOutboxClaimLost
		}
		identity, err := orderModel.FindCreditIdentityForUpdate(txCtx, row.AggregateId)
		if err != nil {
			return err
		}
		if identity.OrderNo != row.AggregateNo {
			return fmt.Errorf(
				"recharge order number mismatch: order_id=%d order_no=%s",
				row.AggregateId,
				row.AggregateNo,
			)
		}
		updated, err = orderModel.MarkCreditRetrying(
			txCtx,
			row.AggregateId,
			identity,
			retryCount,
			now.UnixMilli(),
			truncateError(cause.Error(), 1000),
		)
		if err != nil {
			return err
		}
		if !updated {
			return errors.New("recharge order retry state was not updated")
		}
		return nil
	})
	if err == nil || errors.Is(err, errOutboxClaimLost) {
		return
	}
	// A failed local transaction leaves PROCESSING untouched. Another replica
	// can safely recover it after the lease expires.
	logx.WithContext(ctx).Errorf(
		"persist payment outbox failure failed: id=%d owner=%s err=%v original=%v",
		row.Id,
		owner,
		err,
		cause,
	)
}

func retryDelay(retryCount int64) time.Duration {
	if retryCount <= 0 {
		return time.Second
	}
	delay := time.Duration(retryCount) * time.Second
	if delay > maximumRetryDelay {
		return maximumRetryDelay
	}
	return delay
}

func truncateError(message string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(message)
	if len(runes) <= limit {
		return message
	}
	return string(runes[:limit])
}
