package logic

import (
	"context"
	"fmt"
	"wklive/common/conv"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"
)

func ToAssetStatus(status int64) asset.AssetStatus {
	switch status {
	case 1:
		return asset.AssetStatus_ASSET_STATUS_ENABLED
	case 0:
		return asset.AssetStatus_ASSET_STATUS_DISABLED
	default:
		return asset.AssetStatus_ASSET_STATUS_UNSPECIFIED
	}
}

func ToFreezeStatus(status int64) asset.FreezeStatus {
	switch status {
	case 1:
		return asset.FreezeStatus_FREEZE_STATUS_FREEZING
	case 2:
		return asset.FreezeStatus_FREEZE_STATUS_PARTIAL_RELEASED
	case 3:
		return asset.FreezeStatus_FREEZE_STATUS_UNFROZEN
	case 4:
		return asset.FreezeStatus_FREEZE_STATUS_DEDUCTED
	case 5:
		return asset.FreezeStatus_FREEZE_STATUS_CLOSED
	default:
		return asset.FreezeStatus_FREEZE_STATUS_UNSPECIFIED
	}
}

func ToLockStatus(status int64) asset.LockStatus {
	switch status {
	case 1:
		return asset.LockStatus_LOCK_STATUS_LOCKING
	case 2:
		return asset.LockStatus_LOCK_STATUS_PARTIAL_UNLOCKED
	case 3:
		return asset.LockStatus_LOCK_STATUS_UNLOCKED
	case 4:
		return asset.LockStatus_LOCK_STATUS_CLOSED
	default:
		return asset.LockStatus_LOCK_STATUS_UNSPECIFIED
	}
}

func ToBizTypeValue(bizType string) asset.BizType {
	switch bizType {
	case "payment":
		return asset.BizType_BIZ_TYPE_PAYMENT
	case "trade":
		return asset.BizType_BIZ_TYPE_TRADE
	case "staking":
		return asset.BizType_BIZ_TYPE_STAKING
	case "option":
		return asset.BizType_BIZ_TYPE_OPTION
	case "transfer":
		return asset.BizType_BIZ_TYPE_TRANSFER
	case "system":
		return asset.BizType_BIZ_TYPE_SYSTEM
	case "activity":
		return asset.BizType_BIZ_TYPE_ACTIVITY
	case "earn":
		return asset.BizType_BIZ_TYPE_EARN
	default:
		return asset.BizType_BIZ_TYPE_UNSPECIFIED
	}
}

func ToSceneTypeValue(sceneType string) asset.SceneType {
	switch sceneType {
	case "recharge":
		return asset.SceneType_SCENE_TYPE_RECHARGE
	case "withdraw_apply":
		return asset.SceneType_SCENE_TYPE_WITHDRAW_APPLY
	case "withdraw_reject":
		return asset.SceneType_SCENE_TYPE_WITHDRAW_REJECT
	case "withdraw_success":
		return asset.SceneType_SCENE_TYPE_WITHDRAW_SUCCESS
	case "place_order":
		return asset.SceneType_SCENE_TYPE_PLACE_ORDER
	case "cancel_order":
		return asset.SceneType_SCENE_TYPE_CANCEL_ORDER
	case "trade_match":
		return asset.SceneType_SCENE_TYPE_TRADE_MATCH
	case "trade_fee":
		return asset.SceneType_SCENE_TYPE_TRADE_FEE
	case "staking_join":
		return asset.SceneType_SCENE_TYPE_STAKING_JOIN
	case "staking_release":
		return asset.SceneType_SCENE_TYPE_STAKING_RELEASE
	case "staking_reward":
		return asset.SceneType_SCENE_TYPE_STAKING_REWARD
	case "transfer":
		return asset.SceneType_SCENE_TYPE_TRANSFER
	case "manual_add":
		return asset.SceneType_SCENE_TYPE_MANUAL_ADD
	case "manual_sub":
		return asset.SceneType_SCENE_TYPE_MANUAL_SUB
	case "system_adjust":
		return asset.SceneType_SCENE_TYPE_SYSTEM_ADJUST
	case "airdrop":
		return asset.SceneType_SCENE_TYPE_AIRDROP
	default:
		return asset.SceneType_SCENE_TYPE_UNSPECIFIED
	}
}

func FromBizTypeEnum(bizType asset.BizType) string {
	switch bizType {
	case asset.BizType_BIZ_TYPE_PAYMENT:
		return "payment"
	case asset.BizType_BIZ_TYPE_TRADE:
		return "trade"
	case asset.BizType_BIZ_TYPE_STAKING:
		return "staking"
	case asset.BizType_BIZ_TYPE_OPTION:
		return "option"
	case asset.BizType_BIZ_TYPE_TRANSFER:
		return "transfer"
	case asset.BizType_BIZ_TYPE_SYSTEM:
		return "system"
	case asset.BizType_BIZ_TYPE_ACTIVITY:
		return "activity"
	case asset.BizType_BIZ_TYPE_EARN:
		return "earn"
	default:
		return ""
	}
}

func FromSceneTypeEnum(sceneType asset.SceneType) string {
	switch sceneType {
	case asset.SceneType_SCENE_TYPE_RECHARGE:
		return "recharge"
	case asset.SceneType_SCENE_TYPE_WITHDRAW_APPLY:
		return "withdraw_apply"
	case asset.SceneType_SCENE_TYPE_WITHDRAW_REJECT:
		return "withdraw_reject"
	case asset.SceneType_SCENE_TYPE_WITHDRAW_SUCCESS:
		return "withdraw_success"
	case asset.SceneType_SCENE_TYPE_PLACE_ORDER:
		return "place_order"
	case asset.SceneType_SCENE_TYPE_CANCEL_ORDER:
		return "cancel_order"
	case asset.SceneType_SCENE_TYPE_TRADE_MATCH:
		return "trade_match"
	case asset.SceneType_SCENE_TYPE_TRADE_FEE:
		return "trade_fee"
	case asset.SceneType_SCENE_TYPE_STAKING_JOIN:
		return "staking_join"
	case asset.SceneType_SCENE_TYPE_STAKING_RELEASE:
		return "staking_release"
	case asset.SceneType_SCENE_TYPE_STAKING_REWARD:
		return "staking_reward"
	case asset.SceneType_SCENE_TYPE_TRANSFER:
		return "transfer"
	case asset.SceneType_SCENE_TYPE_MANUAL_ADD:
		return "manual_add"
	case asset.SceneType_SCENE_TYPE_MANUAL_SUB:
		return "manual_sub"
	case asset.SceneType_SCENE_TYPE_SYSTEM_ADJUST:
		return "system_adjust"
	case asset.SceneType_SCENE_TYPE_AIRDROP:
		return "airdrop"
	default:
		return ""
	}
}

func toUserAssetProto(data *models.TUserAsset) *asset.UserAsset {
	if data == nil {
		return nil
	}
	return &asset.UserAsset{
		Id:              data.Id,
		TenantId:        data.TenantId,
		UserId:          data.UserId,
		WalletType:      asset.WalletType(data.WalletType),
		Coin:            data.Coin,
		TotalAmount:     conv.FloatString(data.TotalAmount),
		AvailableAmount: conv.FloatString(data.AvailableAmount),
		FrozenAmount:    conv.FloatString(data.FrozenAmount),
		LockedAmount:    conv.FloatString(data.LockedAmount),
		Status:          ToAssetStatus(data.Status),
		Version:         data.Version,
		Remark:          data.Remark,
		CreateTimes:     data.CreateTimes,
		UpdateTimes:     data.UpdateTimes,
	}
}

func toAssetFlowProto(data *models.TAssetFlow) *asset.AssetFlow {
	if data == nil {
		return nil
	}
	return &asset.AssetFlow{
		Id:                     data.Id,
		FlowNo:                 data.FlowNo,
		TenantId:               data.TenantId,
		UserId:                 data.UserId,
		WalletType:             asset.WalletType(data.WalletType),
		Coin:                   data.Coin,
		BizType:                ToBizTypeValue(data.BizType),
		SceneType:              ToSceneTypeValue(data.SceneType),
		OpType:                 asset.AssetOpType(data.OpType),
		BizId:                  data.BizId,
		BizNo:                  data.BizNo,
		ChangeAmount:           conv.FloatString(data.ChangeAmount),
		BeforeTotalAmount:      conv.FloatString(data.BeforeTotalAmount),
		AfterTotalAmount:       conv.FloatString(data.AfterTotalAmount),
		BeforeAvailableAmount:  conv.FloatString(data.BeforeAvailableAmount),
		AfterAvailableAmount:   conv.FloatString(data.AfterAvailableAmount),
		BeforeFrozenAmount:     conv.FloatString(data.BeforeFrozenAmount),
		AfterFrozenAmount:      conv.FloatString(data.AfterFrozenAmount),
		BeforeLockedAmount:     conv.FloatString(data.BeforeLockedAmount),
		AfterLockedAmount:      conv.FloatString(data.AfterLockedAmount),
		BalanceSnapshotVersion: data.BalanceSnapshotVersion,
		ChangeType:             data.ChangeType,
		Remark:                 data.Remark,
		CreateTimes:            data.CreateTimes,
		UpdateTimes:            data.UpdateTimes,
	}
}

func toAssetFreezeProto(data *models.TAssetFreeze) *asset.AssetFreeze {
	if data == nil {
		return nil
	}
	return &asset.AssetFreeze{
		Id:             data.Id,
		FreezeNo:       data.FreezeNo,
		TenantId:       data.TenantId,
		UserId:         data.UserId,
		WalletType:     asset.WalletType(data.WalletType),
		Coin:           data.Coin,
		BizType:        ToBizTypeValue(data.BizType),
		SceneType:      ToSceneTypeValue(data.SceneType),
		BizId:          data.BizId,
		BizNo:          data.BizNo,
		Amount:         conv.FloatString(data.Amount),
		UsedAmount:     conv.FloatString(data.UsedAmount),
		UnfreezeAmount: conv.FloatString(data.UnfreezeAmount),
		RemainAmount:   conv.FloatString(data.RemainAmount),
		Status:         ToFreezeStatus(data.Status),
		ExpireTime:     data.ExpireTime,
		Remark:         data.Remark,
		CreateTimes:    data.CreateTimes,
		UpdateTimes:    data.UpdateTimes,
	}
}

func toAssetLockProto(data *models.TAssetLock) *asset.AssetLock {
	if data == nil {
		return nil
	}
	return &asset.AssetLock{
		Id:           data.Id,
		LockNo:       data.LockNo,
		TenantId:     data.TenantId,
		UserId:       data.UserId,
		WalletType:   asset.WalletType(data.WalletType),
		Coin:         data.Coin,
		BizType:      ToBizTypeValue(data.BizType),
		SceneType:    ToSceneTypeValue(data.SceneType),
		BizId:        data.BizId,
		BizNo:        data.BizNo,
		Amount:       conv.FloatString(data.Amount),
		UnlockAmount: conv.FloatString(data.UnlockAmount),
		RemainAmount: conv.FloatString(data.RemainAmount),
		Status:       ToLockStatus(data.Status),
		StartTime:    data.StartTime,
		EndTime:      data.EndTime,
		Remark:       data.Remark,
		CreateTimes:  data.CreateTimes,
		UpdateTimes:  data.UpdateTimes,
	}
}

func buildAssetFlowRecord(svcCtx *svc.ServiceContext, ctx context.Context, tenantId, userId, walletType int64, coin, changeType, bizType, sceneType string, bizId int64, bizNo string, opType asset.AssetOpType, amount float64, before *models.TUserAsset, after *models.TUserAsset, remark string, ts int64) *models.TAssetFlow {
	beforeTotal, beforeAvailable, beforeFrozen, beforeLocked := 0.0, 0.0, 0.0, 0.0
	afterTotal, afterAvailable, afterFrozen, afterLocked := 0.0, 0.0, 0.0, 0.0
	if before != nil {
		beforeTotal = before.TotalAmount
		beforeAvailable = before.AvailableAmount
		beforeFrozen = before.FrozenAmount
		beforeLocked = before.LockedAmount
	}
	if after != nil {
		afterTotal = after.TotalAmount
		afterAvailable = after.AvailableAmount
		afterFrozen = after.FrozenAmount
		afterLocked = after.LockedAmount
	}
	flowNo, err := svcCtx.GenerateOrderNo(ctx, "FLOW", bizNo)
	if err != nil {
		return nil
	}
	return &models.TAssetFlow{
		FlowNo:                 flowNo,
		TenantId:               tenantId,
		UserId:                 userId,
		WalletType:             walletType,
		Coin:                   coin,
		ChangeType:             changeType,
		BizType:                bizType,
		SceneType:              sceneType,
		BizId:                  bizId,
		BizNo:                  bizNo,
		OpType:                 int64(opType),
		ChangeAmount:           amount,
		BeforeTotalAmount:      beforeTotal,
		AfterTotalAmount:       afterTotal,
		BeforeAvailableAmount:  beforeAvailable,
		AfterAvailableAmount:   afterAvailable,
		BeforeFrozenAmount:     beforeFrozen,
		AfterFrozenAmount:      afterFrozen,
		BeforeLockedAmount:     beforeLocked,
		AfterLockedAmount:      afterLocked,
		BalanceSnapshotVersion: 0,
		Remark:                 remark,
		CreateTimes:            ts,
		UpdateTimes:            ts,
	}
}

func buildAssetFreezeRecord(svcCtx *svc.ServiceContext, ctx context.Context, tenantId, userId, walletType int64, coin, bizType, sceneType, bizNo, remark string, amount float64, expireTime, ts int64) *models.TAssetFreeze {
	freezeNo, err := svcCtx.GenerateOrderNo(ctx, "FREEZE", bizNo)
	if err != nil {
		return nil
	}
	return &models.TAssetFreeze{
		FreezeNo:       freezeNo,
		TenantId:       tenantId,
		UserId:         userId,
		WalletType:     walletType,
		Coin:           coin,
		BizType:        bizType,
		SceneType:      sceneType,
		BizNo:          bizNo,
		Amount:         amount,
		UsedAmount:     0,
		UnfreezeAmount: 0,
		RemainAmount:   amount,
		Status:         1,
		ExpireTime:     expireTime,
		Remark:         remark,
		CreateTimes:    ts,
		UpdateTimes:    ts,
	}
}

func buildAssetLockRecord(svcCtx *svc.ServiceContext, ctx context.Context, tenantId, userId, walletType int64, coin, bizType, sceneType, bizNo, remark string, amount float64, startTime, endTime, ts int64) *models.TAssetLock {
	lockNo, err := svcCtx.GenerateOrderNo(ctx, "FREEZE", bizNo)
	if err != nil {
		return nil
	}
	return &models.TAssetLock{
		LockNo:       lockNo,
		TenantId:     tenantId,
		UserId:       userId,
		WalletType:   walletType,
		Coin:         coin,
		BizType:      bizType,
		SceneType:    sceneType,
		BizNo:        bizNo,
		Amount:       amount,
		UnlockAmount: 0,
		RemainAmount: amount,
		Status:       1,
		StartTime:    startTime,
		EndTime:      endTime,
		Remark:       remark,
		CreateTimes:  ts,
		UpdateTimes:  ts,
	}
}

func assetBizType(in asset.BizType) string {
	return FromBizTypeEnum(in)
}

func assetSceneType(in asset.SceneType) string {
	return FromSceneTypeEnum(in)
}

func assetStatusFilter(status asset.AssetStatus) int64 {
	switch status {
	case asset.AssetStatus_ASSET_STATUS_ENABLED:
		return 1
	case asset.AssetStatus_ASSET_STATUS_DISABLED:
		return 0
	default:
		return 0
	}
}

func EnumToFilterString(bizType asset.BizType, sceneType asset.SceneType) (string, string) {
	return assetBizType(bizType), assetSceneType(sceneType)
}

func findFreezeByBizNo(ctx context.Context, svcCtx *svc.ServiceContext, tenantId int64, bizType asset.BizType, bizNo string) (*models.TAssetFreeze, error) {
	list, _, err := svcCtx.AssetFreezeModel.FindPage(ctx, tenantId, 0, 0, "", FromBizTypeEnum(bizType), bizNo, 0, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("freeze record not found for bizNo=%s", bizNo)
	}
	return list[0], nil
}

func findLockByBizNo(ctx context.Context, svcCtx *svc.ServiceContext, tenantId int64, bizType asset.BizType, bizNo string) (*models.TAssetLock, error) {
	list, _, err := svcCtx.AssetLockModel.FindPage(ctx, tenantId, 0, 0, "", FromBizTypeEnum(bizType), bizNo, 0, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("lock record not found for bizNo=%s", bizNo)
	}
	return list[0], nil
}
