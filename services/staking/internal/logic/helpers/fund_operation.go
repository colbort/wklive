package helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/delayqueue"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrStakeOperationProcessing = errors.New("staking operation is processing")

type RedeemOperationSpec struct {
	OperationNo   string
	RequestNo     string
	OperationType int64
	RedeemType    staking.RedeemType
	RewardAmount  decimal.Decimal
	FeeRate       decimal.Decimal
	FeeAmount     decimal.Decimal
	OperatorId    int64
	Remark        string
}

type RewardOperationSpec struct {
	OperationNo   string
	RequestNo     string
	OperationType int64
	RewardType    staking.RewardType
	RewardAmount  decimal.Decimal
	PeriodEnd     int64
	OperatorId    int64
	Remark        string
}

func ExecuteSubscribeOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	if operation.Status == StakeOperationStatusSucceeded {
		return nil
	}
	now := utils.NowMillis()
	claimed, err := svcCtx.StakeOperationModel.Claim(ctx, operation.Id, now)
	if err != nil {
		return err
	}
	if !claimed {
		return ErrStakeOperationProcessing
	}
	operation.Status = StakeOperationStatusProcessing
	operation.Version++
	operation.UpdateTimes = now
	order, err := svcCtx.StakeOrderModel.FindOne(ctx, operation.OrderId)
	if err != nil {
		return markRedeemOperationFailed(ctx, svcCtx, operation, err)
	}
	if order.Status != int64(staking.OrderStatus_ORDER_STATUS_PENDING) {
		return markRedeemOperationFailed(ctx, svcCtx, operation, fmt.Errorf("subscribe order status is %d", order.Status))
	}
	if operation.PrincipalStatus != StakeOperationStepSucceeded {
		resp, callErr := svcCtx.AssetClient.LockAsset(ctx, &asset.LockAssetReq{
			TenantId: order.TenantId, UserId: order.UserId, WalletType: common.WalletType_WALLET_TYPE_EARN,
			Coin: order.CoinSymbol, Amount: conv.FloatString(order.StakeAmount),
			BizType: asset.BizType_BIZ_TYPE_STAKING, SceneType: asset.SceneType_SCENE_TYPE_STAKING_JOIN,
			BizId: order.Id, BizNo: order.OrderNo, StartTime: order.StartTimes, EndTime: order.EndTimes, Remark: operation.Remark,
		})
		if callErr != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, callErr)
		}
		if err := requireAssetSuccess(resp.GetBase()); err != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, err)
		}
		operation.PrincipalStatus = StakeOperationStepSucceeded
		if err := persistProcessingOperation(ctx, svcCtx, operation); err != nil {
			return err
		}
	}

	now = utils.NowMillis()
	err = svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTStakeOrderModel(conn, svcCtx.Config.CacheRedis)
		operationModel := models.NewTStakeOperationModel(conn, svcCtx.Config.CacheRedis)
		currentOperation, err := lockOwnedStakeOperation(ctx, operationModel, operation)
		if err != nil {
			return err
		}
		current, err := orderModel.FindOneForUpdate(ctx, order.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(staking.OrderStatus_ORDER_STATUS_PENDING) {
			current.Status = int64(staking.OrderStatus_ORDER_STATUS_STAKING)
			current.Version++
			current.UpdateTimes = now
			if err := orderModel.Update(ctx, current); err != nil {
				return err
			}
		}
		return completeStakeOperation(ctx, operationModel, currentOperation, operation, now)
	})
	if err != nil {
		return err
	}
	if order.EndTimes > 0 && svcCtx.DelayQueue != nil {
		_ = svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionMatureOrder, TenantID: order.TenantId, OrderID: order.Id, DueAt: order.EndTimes,
		}, time.UnixMilli(order.EndTimes))
	}
	return nil
}

func PrepareRewardOperation(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TStakeOrder, spec RewardOperationSpec) (*models.TStakeOperation, error) {
	spec.RequestNo = strings.TrimSpace(spec.RequestNo)
	if spec.RequestNo == "" || len(spec.RequestNo) > 96 || spec.OperationNo == "" || len(spec.OperationNo) > 96 || !spec.RewardAmount.IsPositive() {
		return nil, i18n.StatusError(ctx, i18n.ParamError)
	}
	if existing, err := svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(ctx, order.TenantId, order.UserId, spec.OperationType, spec.RequestNo); err == nil {
		if existing.OrderId != order.Id || !existing.RewardAmount.Equal(spec.RewardAmount) {
			return nil, i18n.StatusError(ctx, i18n.ParamError)
		}
		log, logErr := svcCtx.StakeRewardLogModel.FindOneByTenantIdOperationNo(ctx, order.TenantId, existing.OperationNo)
		if logErr != nil {
			return nil, logErr
		}
		if log.RewardType != int64(spec.RewardType) {
			return nil, i18n.StatusError(ctx, i18n.ParamError)
		}
		return existing, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	now := utils.NowMillis()
	operation := &models.TStakeOperation{
		TenantId: order.TenantId, UserId: order.UserId, OrderId: order.Id, OrderNo: order.OrderNo,
		OperationNo: spec.OperationNo, RequestNo: spec.RequestNo, OperationType: spec.OperationType,
		PrincipalAmount: decimal.Zero, RewardAmount: spec.RewardAmount, FeeAmount: decimal.Zero,
		PrincipalStatus: StakeOperationStepNotRequired, RewardStatus: StakeOperationStepPending, FeeStatus: StakeOperationStepNotRequired,
		Status: StakeOperationStatusPending, PeriodEnd: spec.PeriodEnd, OperatorUserId: spec.OperatorId,
		Remark: spec.Remark, Version: 1, CreateTimes: now, UpdateTimes: now,
	}
	err := svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		operationModel := models.NewTStakeOperationModel(conn, svcCtx.Config.CacheRedis)
		rewardLogModel := models.NewTStakeRewardLogModel(conn, svcCtx.Config.CacheRedis)
		if _, err := operationModel.Insert(ctx, operation); err != nil {
			return err
		}
		_, err := rewardLogModel.Insert(ctx, &models.TStakeRewardLog{
			TenantId: order.TenantId, OrderId: order.Id, OrderNo: order.OrderNo, OperationNo: spec.OperationNo, UserId: order.UserId,
			ProductId: order.ProductId, ProductName: order.ProductName, CoinSymbol: order.CoinSymbol,
			RewardCoinSymbol: order.RewardCoinSymbol, RewardAmount: spec.RewardAmount,
			BeforeReward: order.TotalReward, AfterReward: order.TotalReward,
			RewardType: int64(spec.RewardType), RewardStatus: int64(staking.RewardStatus_REWARD_STATUS_PROCESSING),
			RewardTimes: 0, Remark: spec.Remark,
			CreateUserId: spec.OperatorId, UpdateUserId: spec.OperatorId, CreateTimes: now, UpdateTimes: now,
		})
		return err
	})
	if err != nil {
		if existing, findErr := svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(ctx, order.TenantId, order.UserId, spec.OperationType, spec.RequestNo); findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return svcCtx.StakeOperationModel.FindOneByTenantIdOperationNo(ctx, order.TenantId, spec.OperationNo)
}

func ExecuteRewardOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	if operation.Status == StakeOperationStatusSucceeded {
		return nil
	}
	now := utils.NowMillis()
	claimed, err := svcCtx.StakeOperationModel.Claim(ctx, operation.Id, now)
	if err != nil {
		return err
	}
	if !claimed {
		return ErrStakeOperationProcessing
	}
	operation.Status = StakeOperationStatusProcessing
	operation.Version++
	operation.UpdateTimes = now
	order, err := svcCtx.StakeOrderModel.FindOne(ctx, operation.OrderId)
	if err != nil {
		return markRedeemOperationFailed(ctx, svcCtx, operation, err)
	}
	if operation.RewardStatus != StakeOperationStepSucceeded {
		resp, callErr := svcCtx.AssetClient.PayPlatformExpense(ctx, &asset.PayPlatformExpenseReq{
			TenantId: order.TenantId, UserId: order.UserId, WalletType: common.WalletType_WALLET_TYPE_EARN,
			PlatformAccountType: "STAKING_REWARD", Coin: order.RewardCoinSymbol, Amount: conv.FloatString(operation.RewardAmount),
			BizType: asset.BizType_BIZ_TYPE_STAKING, SceneType: asset.SceneType_SCENE_TYPE_STAKING_REWARD,
			BizId: order.Id, BizNo: StakeOperationStepBizNo(operation.OperationNo, "reward"), Remark: operation.Remark,
		})
		if callErr != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, callErr)
		}
		if err := requireAssetSuccess(resp.GetBase()); err != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, err)
		}
		operation.RewardStatus = StakeOperationStepSucceeded
		if err := persistProcessingOperation(ctx, svcCtx, operation); err != nil {
			return err
		}
	}
	return finalizeRewardOperation(ctx, svcCtx, operation)
}

func finalizeRewardOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	now := utils.NowMillis()
	return svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTStakeOrderModel(conn, svcCtx.Config.CacheRedis)
		operationModel := models.NewTStakeOperationModel(conn, svcCtx.Config.CacheRedis)
		rewardLogModel := models.NewTStakeRewardLogModel(conn, svcCtx.Config.CacheRedis)
		currentOperation, err := lockOwnedStakeOperation(ctx, operationModel, operation)
		if err != nil {
			return err
		}
		order, err := orderModel.FindOneForUpdate(ctx, operation.OrderId)
		if err != nil {
			return err
		}
		rewardLog, err := rewardLogModel.FindOneByTenantIdOperationNo(ctx, operation.TenantId, operation.OperationNo)
		if err != nil {
			return err
		}
		before := order.TotalReward
		order.TotalReward = order.TotalReward.Add(operation.RewardAmount)
		order.LastRewardTimes = now
		if operation.PeriodEnd > 0 {
			order.LastRewardTimes = operation.PeriodEnd
			order.InterestDays++
			order.NextRewardTimes = operation.PeriodEnd + 24*60*60*1000
		}
		order.UpdateUserId = operation.OperatorUserId
		order.UpdateTimes = now
		order.Version++
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		rewardLog.BeforeReward = before
		rewardLog.AfterReward = order.TotalReward
		rewardLog.RewardStatus = int64(staking.RewardStatus_REWARD_STATUS_SUCCESS)
		rewardLog.RewardTimes = now
		rewardLog.UpdateTimes = now
		if err := rewardLogModel.Update(ctx, rewardLog); err != nil {
			return err
		}
		return completeStakeOperation(ctx, operationModel, currentOperation, operation, now)
	})
}

func PrepareRedeemOperation(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TStakeOrder, spec RedeemOperationSpec) (*models.TStakeOperation, error) {
	spec.RequestNo = strings.TrimSpace(spec.RequestNo)
	if spec.RequestNo == "" || len(spec.RequestNo) > 96 || spec.OperationNo == "" || len(spec.OperationNo) > 96 {
		return nil, i18n.StatusError(ctx, i18n.ParamError)
	}
	if existing, err := svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(ctx, order.TenantId, order.UserId, spec.OperationType, spec.RequestNo); err == nil {
		if existing.OrderId != order.Id || !existing.PrincipalAmount.Equal(order.StakeAmount) ||
			!existing.RewardAmount.Equal(spec.RewardAmount) || !existing.FeeAmount.Equal(spec.FeeAmount) {
			return nil, i18n.StatusError(ctx, i18n.ParamError)
		}
		log, logErr := svcCtx.StakeRedeemLogModel.FindOneByTenantIdRedeemNo(ctx, order.TenantId, existing.OperationNo)
		if logErr != nil {
			return nil, logErr
		}
		if log.RedeemType != int64(spec.RedeemType) || !log.FeeRate.Equal(spec.FeeRate) {
			return nil, i18n.StatusError(ctx, i18n.ParamError)
		}
		return existing, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	now := utils.NowMillis()
	operation := &models.TStakeOperation{
		TenantId:        order.TenantId,
		UserId:          order.UserId,
		OrderId:         order.Id,
		OrderNo:         order.OrderNo,
		OperationNo:     spec.OperationNo,
		RequestNo:       spec.RequestNo,
		OperationType:   spec.OperationType,
		PrincipalAmount: order.StakeAmount,
		RewardAmount:    spec.RewardAmount,
		FeeAmount:       spec.FeeAmount,
		PrincipalStatus: StakeOperationStepPending,
		RewardStatus:    stepStatus(spec.RewardAmount),
		FeeStatus:       stepStatus(spec.FeeAmount),
		Status:          StakeOperationStatusPending,
		OperatorUserId:  spec.OperatorId,
		Remark:          spec.Remark,
		Version:         1,
		CreateTimes:     now,
		UpdateTimes:     now,
	}

	err := svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTStakeOrderModel(conn, svcCtx.Config.CacheRedis)
		operationModel := models.NewTStakeOperationModel(conn, svcCtx.Config.CacheRedis)
		redeemLogModel := models.NewTStakeRedeemLogModel(conn, svcCtx.Config.CacheRedis)

		claimed, err := orderModel.ClaimOperation(ctx, order.Id, spec.OperationNo, now, []int64{
			int64(staking.OrderStatus_ORDER_STATUS_STAKING),
			int64(staking.OrderStatus_ORDER_STATUS_EXPIRED),
		})
		if err != nil {
			return err
		}
		if !claimed {
			return ErrStakeOperationProcessing
		}
		if _, err := operationModel.Insert(ctx, operation); err != nil {
			return err
		}
		_, err = redeemLogModel.Insert(ctx, &models.TStakeRedeemLog{
			TenantId:     order.TenantId,
			OrderId:      order.Id,
			OrderNo:      order.OrderNo,
			UserId:       order.UserId,
			ProductId:    order.ProductId,
			RedeemNo:     spec.OperationNo,
			RedeemType:   int64(spec.RedeemType),
			StakeAmount:  order.StakeAmount,
			RedeemAmount: order.StakeAmount,
			RewardAmount: spec.RewardAmount,
			FeeRate:      spec.FeeRate,
			FeeAmount:    spec.FeeAmount,
			RedeemStatus: int64(staking.RedeemStatus_REDEEM_STATUS_PROCESSING),
			RedeemTimes:  0,
			Remark:       spec.Remark,
			CreateUserId: spec.OperatorId,
			UpdateUserId: spec.OperatorId,
			CreateTimes:  now,
			UpdateTimes:  now,
		})
		return err
	})
	if err != nil {
		if existing, findErr := svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(ctx, order.TenantId, order.UserId, spec.OperationType, spec.RequestNo); findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return svcCtx.StakeOperationModel.FindOneByTenantIdOperationNo(ctx, order.TenantId, spec.OperationNo)
}

func ExecuteRedeemOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	if operation.Status == StakeOperationStatusSucceeded {
		return nil
	}
	now := utils.NowMillis()
	claimed, err := svcCtx.StakeOperationModel.Claim(ctx, operation.Id, now)
	if err != nil {
		return err
	}
	if !claimed {
		return ErrStakeOperationProcessing
	}
	operation.Status = StakeOperationStatusProcessing
	operation.Version++
	operation.UpdateTimes = now

	// ClaimOperation updates active_operation_no with a conditional SQL update.
	// Read outside cache so a stale pre-claim value cannot make a valid durable
	// operation fail with "active operation mismatch".
	order, err := svcCtx.StakeOrderModel.FindOneForUpdate(ctx, operation.OrderId)
	if err != nil {
		return markRedeemOperationFailed(ctx, svcCtx, operation, err)
	}
	if order.ActiveOperationNo != operation.OperationNo {
		return markRedeemOperationFailed(ctx, svcCtx, operation, fmt.Errorf("order active operation mismatch"))
	}

	unlockAmount := operation.PrincipalAmount.Sub(operation.FeeAmount)
	if unlockAmount.IsNegative() {
		unlockAmount = decimal.Zero
	}
	if operation.PrincipalStatus != StakeOperationStepSucceeded && unlockAmount.IsPositive() {
		resp, callErr := svcCtx.AssetClient.UnlockAssetByBizNo(ctx, &asset.UnlockAssetByBizNoReq{
			TenantId:      order.TenantId,
			TargetBizType: asset.BizType_BIZ_TYPE_STAKING,
			TargetBizNo:   order.OrderNo,
			Amount:        conv.FloatString(unlockAmount),
			BizType:       asset.BizType_BIZ_TYPE_STAKING,
			SceneType:     asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
			BizId:         order.Id,
			BizNo:         StakeOperationStepBizNo(operation.OperationNo, "principal"),
			Remark:        operation.Remark,
		})
		if callErr != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, callErr)
		}
		if err := requireAssetSuccess(resp.GetBase()); err != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, err)
		}
		operation.PrincipalStatus = StakeOperationStepSucceeded
		if err := persistProcessingOperation(ctx, svcCtx, operation); err != nil {
			return err
		}
	} else if !unlockAmount.IsPositive() {
		operation.PrincipalStatus = StakeOperationStepSucceeded
	}

	if operation.FeeStatus != StakeOperationStepSucceeded && operation.FeeAmount.IsPositive() {
		resp, callErr := svcCtx.AssetClient.CollectLockedRevenue(ctx, &asset.CollectLockedRevenueReq{
			TenantId:            order.TenantId,
			TargetBizType:       asset.BizType_BIZ_TYPE_STAKING,
			TargetBizNo:         order.OrderNo,
			PlatformAccountType: "FEE_REVENUE",
			Amount:              conv.FloatString(operation.FeeAmount),
			BizType:             asset.BizType_BIZ_TYPE_STAKING,
			SceneType:           asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
			BizId:               order.Id,
			BizNo:               StakeOperationStepBizNo(operation.OperationNo, "fee"),
			Remark:              "staking redeem fee",
		})
		if callErr != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, callErr)
		}
		if err := requireAssetSuccess(resp.GetBase()); err != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, err)
		}
		operation.FeeStatus = StakeOperationStepSucceeded
		if err := persistProcessingOperation(ctx, svcCtx, operation); err != nil {
			return err
		}
	}

	if operation.RewardStatus != StakeOperationStepSucceeded && operation.RewardAmount.IsPositive() {
		resp, callErr := svcCtx.AssetClient.PayPlatformExpense(ctx, &asset.PayPlatformExpenseReq{
			TenantId:            order.TenantId,
			UserId:              order.UserId,
			WalletType:          common.WalletType_WALLET_TYPE_EARN,
			PlatformAccountType: "STAKING_REWARD",
			Coin:                order.RewardCoinSymbol,
			Amount:              conv.FloatString(operation.RewardAmount),
			BizType:             asset.BizType_BIZ_TYPE_STAKING,
			SceneType:           asset.SceneType_SCENE_TYPE_STAKING_REWARD,
			BizId:               order.Id,
			BizNo:               StakeOperationStepBizNo(operation.OperationNo, "reward"),
			Remark:              operation.Remark,
		})
		if callErr != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, callErr)
		}
		if err := requireAssetSuccess(resp.GetBase()); err != nil {
			return markRedeemOperationFailed(ctx, svcCtx, operation, err)
		}
		operation.RewardStatus = StakeOperationStepSucceeded
		if err := persistProcessingOperation(ctx, svcCtx, operation); err != nil {
			return err
		}
	}

	return finalizeRedeemOperation(ctx, svcCtx, operation)
}

func finalizeRedeemOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	now := utils.NowMillis()
	return svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTStakeOrderModel(conn, svcCtx.Config.CacheRedis)
		productModel := models.NewTStakeProductModel(conn, svcCtx.Config.CacheRedis)
		positionModel := models.NewTStakeUserPositionModel(conn, svcCtx.Config.CacheRedis)
		operationModel := models.NewTStakeOperationModel(conn, svcCtx.Config.CacheRedis)
		redeemLogModel := models.NewTStakeRedeemLogModel(conn, svcCtx.Config.CacheRedis)
		currentOperation, err := lockOwnedStakeOperation(ctx, operationModel, operation)
		if err != nil {
			return err
		}

		order, err := orderModel.FindOneForUpdate(ctx, operation.OrderId)
		if err != nil {
			return err
		}
		if order.ActiveOperationNo != operation.OperationNo {
			return fmt.Errorf("order active operation changed")
		}
		redeemLog, err := redeemLogModel.FindOneByTenantIdRedeemNo(ctx, operation.TenantId, operation.OperationNo)
		if err != nil {
			return err
		}
		if err := productModel.ReleaseStakeAmount(ctx, order.ProductId, order.StakeAmount, now); err != nil {
			return err
		}
		if err := positionModel.ReleaseAmount(ctx, order.TenantId, order.UserId, order.ProductId, order.StakeAmount, now); err != nil {
			return err
		}

		order.RedeemAmount = order.StakeAmount
		order.RedeemFee = operation.FeeAmount
		order.TotalReward = order.TotalReward.Add(operation.RewardAmount)
		order.PendingReward = decimal.Zero
		order.RedeemType = redeemLog.RedeemType
		order.RedeemTimes = now
		order.ActiveOperationNo = ""
		if shouldMarkEarlyRedeemed(redeemLog.RedeemType, order.EndTimes, now) {
			order.Status = int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED)
		} else {
			order.Status = int64(staking.OrderStatus_ORDER_STATUS_REDEEMED)
		}
		order.Remark = operation.Remark
		order.UpdateUserId = operation.OperatorUserId
		order.UpdateTimes = now
		order.Version++
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}

		redeemLog.RedeemStatus = int64(staking.RedeemStatus_REDEEM_STATUS_SUCCESS)
		redeemLog.RedeemTimes = now
		redeemLog.UpdateUserId = operation.OperatorUserId
		redeemLog.UpdateTimes = now
		if err := redeemLogModel.Update(ctx, redeemLog); err != nil {
			return err
		}

		operation.PrincipalStatus = StakeOperationStepSucceeded
		if !operation.FeeAmount.IsPositive() {
			operation.FeeStatus = StakeOperationStepNotRequired
		}
		if !operation.RewardAmount.IsPositive() {
			operation.RewardStatus = StakeOperationStepNotRequired
		}
		return completeStakeOperation(ctx, operationModel, currentOperation, operation, now)
	})
}

func shouldMarkEarlyRedeemed(redeemType, endTimes, now int64) bool {
	if redeemType == int64(staking.RedeemType_REDEEM_TYPE_EARLY) {
		return true
	}
	return redeemType == int64(staking.RedeemType_REDEEM_TYPE_MANUAL) && (endTimes == 0 || endTimes > now)
}

func stepStatus(amount decimal.Decimal) int64 {
	if amount.IsPositive() {
		return StakeOperationStepPending
	}
	return StakeOperationStepNotRequired
}

func persistProcessingOperation(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation) error {
	now := utils.NowMillis()
	version, err := svcCtx.StakeOperationModel.CheckpointSteps(
		ctx, operation.Id, operation.Version,
		operation.PrincipalStatus, operation.RewardStatus, operation.FeeStatus, now,
	)
	if err != nil {
		return err
	}
	operation.Status = StakeOperationStatusProcessing
	operation.Version = version
	operation.UpdateTimes = now
	return nil
}

func markRedeemOperationFailed(ctx context.Context, svcCtx *svc.ServiceContext, operation *models.TStakeOperation, cause error) error {
	operation.RetryCount++
	now := utils.NowMillis()
	status := StakeOperationRetryStatus(operation.RetryCount)
	next := now + int64((30 * time.Second).Milliseconds())
	if err := svcCtx.StakeOperationModel.MarkRetryable(ctx, operation.Id, operation.Version, operation.RetryCount, next, status, now, truncateOperationError(cause.Error())); err != nil {
		return fmt.Errorf("%w; mark retry failed: %v", cause, err)
	}
	return cause
}

func lockOwnedStakeOperation(
	ctx context.Context, model models.TStakeOperationModel, operation *models.TStakeOperation,
) (*models.TStakeOperation, error) {
	current, err := model.FindOneForUpdate(ctx, operation.Id)
	if err != nil {
		return nil, err
	}
	if current.Status != StakeOperationStatusProcessing || current.Version != operation.Version {
		return nil, ErrStakeOperationProcessing
	}
	return current, nil
}

func completeStakeOperation(
	ctx context.Context,
	model models.TStakeOperationModel,
	current, operation *models.TStakeOperation,
	now int64,
) error {
	current.PrincipalStatus = operation.PrincipalStatus
	current.RewardStatus = operation.RewardStatus
	current.FeeStatus = operation.FeeStatus
	current.Status = StakeOperationStatusSucceeded
	current.NextRetryAt = 0
	current.LastError = ""
	current.Version++
	current.UpdateTimes = now
	return model.Update(ctx, current)
}

func requireAssetSuccess(base *common.RespBase) error {
	if base == nil {
		return errors.New("asset response base is nil")
	}
	if base.Code != 200 {
		return fmt.Errorf("asset rejected operation: code=%d msg=%s", base.Code, base.Msg)
	}
	return nil
}

func truncateOperationError(value string) string {
	if len(value) <= 1000 {
		return value
	}
	return value[:1000]
}
