package logic

import (
	"wklive/common/conv"
	"wklive/proto/staking"
	"wklive/services/staking/models"
)

func productToProto(item *models.TStakeProduct) *staking.StakeProduct {
	if item == nil {
		return nil
	}
	return &staking.StakeProduct{
		Id:               item.Id,
		TenantId:         item.TenantId,
		ProductNo:        item.ProductNo,
		ProductName:      item.ProductName,
		ProductType:      staking.ProductType(item.ProductType),
		CoinName:         item.CoinName,
		CoinSymbol:       item.CoinSymbol,
		RewardCoinName:   item.RewardCoinName,
		RewardCoinSymbol: item.RewardCoinSymbol,
		Apr:              conv.FloatString(item.Apr),
		LockDays:         int32(item.LockDays),
		MinAmount:        conv.FloatString(item.MinAmount),
		MaxAmount:        conv.FloatString(item.MaxAmount),
		StepAmount:       conv.FloatString(item.StepAmount),
		TotalAmount:      conv.FloatString(item.TotalAmount),
		StakedAmount:     conv.FloatString(item.StakedAmount),
		UserLimitAmount:  conv.FloatString(item.UserLimitAmount),
		InterestMode:     staking.InterestMode(item.InterestMode),
		RewardMode:       staking.RewardMode(item.RewardMode),
		AllowEarlyRedeem: staking.YesNo(item.AllowEarlyRedeem),
		EarlyRedeemRate:  conv.FloatString(item.EarlyRedeemRate),
		Status:           staking.ProductStatus(item.Status),
		Sort:             int32(item.Sort),
		Remark:           item.Remark,
		CreateUserId:     item.CreateUserId,
		UpdateUserId:     item.UpdateUserId,
		CreateTimes:      int64(item.CreateTimes),
		UpdateTimes:      int64(item.UpdateTimes),
	}
}

func orderToProto(item *models.TStakeOrder) *staking.StakeOrder {
	if item == nil {
		return nil
	}
	return &staking.StakeOrder{
		Id:               item.Id,
		TenantId:         item.TenantId,
		OrderNo:          item.OrderNo,
		Uid:              item.Uid,
		ProductId:        item.ProductId,
		ProductNo:        item.ProductNo,
		ProductName:      item.ProductName,
		ProductType:      staking.ProductType(item.ProductType),
		CoinName:         item.CoinName,
		CoinSymbol:       item.CoinSymbol,
		RewardCoinName:   item.RewardCoinName,
		RewardCoinSymbol: item.RewardCoinSymbol,
		StakeAmount:      conv.FloatString(item.StakeAmount),
		Apr:              conv.FloatString(item.Apr),
		LockDays:         int32(item.LockDays),
		InterestMode:     staking.InterestMode(item.InterestMode),
		RewardMode:       staking.RewardMode(item.RewardMode),
		AllowEarlyRedeem: staking.YesNo(item.AllowEarlyRedeem),
		EarlyRedeemRate:  conv.FloatString(item.EarlyRedeemRate),
		InterestDays:     int32(item.InterestDays),
		StartTimes:       int64(item.StartTimes),
		EndTimes:         int64(item.EndTimes),
		LastRewardTimes:  int64(item.LastRewardTimes),
		NextRewardTimes:  int64(item.NextRewardTimes),
		TotalReward:      conv.FloatString(item.TotalReward),
		PendingReward:    conv.FloatString(item.PendingReward),
		RedeemAmount:     conv.FloatString(item.RedeemAmount),
		RedeemFee:        conv.FloatString(item.RedeemFee),
		Status:           staking.OrderStatus(item.Status),
		RedeemType:       staking.RedeemType(item.RedeemType),
		RedeemApplyTimes: int64(item.RedeemApplyTimes),
		RedeemTimes:      int64(item.RedeemTimes),
		Source:           staking.SourceType(item.Source),
		Remark:           item.Remark,
		CreateUserId:     item.CreateUserId,
		UpdateUserId:     item.UpdateUserId,
		CreateTimes:      int64(item.CreateTimes),
		UpdateTimes:      int64(item.UpdateTimes),
	}
}

func rewardLogToProto(item *models.TStakeRewardLog) *staking.StakeRewardLog {
	if item == nil {
		return nil
	}
	return &staking.StakeRewardLog{
		Id:               item.Id,
		TenantId:         item.TenantId,
		OrderId:          item.OrderId,
		OrderNo:          item.OrderNo,
		Uid:              item.Uid,
		ProductId:        item.ProductId,
		ProductName:      item.ProductName,
		CoinSymbol:       item.CoinSymbol,
		RewardCoinSymbol: item.RewardCoinSymbol,
		RewardAmount:     conv.FloatString(item.RewardAmount),
		BeforeReward:     conv.FloatString(item.BeforeReward),
		AfterReward:      conv.FloatString(item.AfterReward),
		RewardType:       staking.RewardType(item.RewardType),
		RewardStatus:     staking.RewardStatus(item.RewardStatus),
		RewardTimes:      int64(item.RewardTimes),
		Remark:           item.Remark,
		CreateUserId:     item.CreateUserId,
		UpdateUserId:     item.UpdateUserId,
		CreateTimes:      int64(item.CreateTimes),
		UpdateTimes:      int64(item.UpdateTimes),
	}
}

func redeemLogToProto(item *models.TStakeRedeemLog) *staking.StakeRedeemLog {
	if item == nil {
		return nil
	}
	return &staking.StakeRedeemLog{
		Id:           item.Id,
		TenantId:     item.TenantId,
		OrderId:      item.OrderId,
		OrderNo:      item.OrderNo,
		Uid:          item.Uid,
		ProductId:    item.ProductId,
		RedeemNo:     item.RedeemNo,
		RedeemType:   staking.RedeemType(item.RedeemType),
		StakeAmount:  conv.FloatString(item.StakeAmount),
		RedeemAmount: conv.FloatString(item.RedeemAmount),
		RewardAmount: conv.FloatString(item.RewardAmount),
		FeeRate:      conv.FloatString(item.FeeRate),
		FeeAmount:    conv.FloatString(item.FeeAmount),
		RedeemStatus: staking.RedeemStatus(item.RedeemStatus),
		RedeemTimes:  int64(item.RedeemTimes),
		Remark:       item.Remark,
		CreateUserId: item.CreateUserId,
		UpdateUserId: item.UpdateUserId,
		CreateTimes:  int64(item.CreateTimes),
		UpdateTimes:  int64(item.UpdateTimes),
	}
}

func activeOrderStatuses() []int64 {
	return []int64{
		int64(staking.OrderStatus_ORDER_STATUS_STAKING),
		int64(staking.OrderStatus_ORDER_STATUS_EXPIRED),
	}
}

func calcNextRewardTime(now int64, rewardMode staking.RewardMode, endTime int64) int64 {
	if rewardMode == staking.RewardMode_REWARD_MODE_DAILY {
		return now + 24*3600*1000
	}
	return endTime
}
