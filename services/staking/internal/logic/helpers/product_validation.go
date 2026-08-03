package helpers

import (
	"context"
	"errors"
	"strings"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
)

const (
	StakingRewardAccountType = "STAKING_REWARD"
	FeeRevenueAccountType    = "FEE_REVENUE"
)

var ErrInvalidStakeProduct = errors.New("invalid staking product configuration")

// ValidateStakeProduct keeps invalid economic parameters out of the order
// path. APR and early redeem rate use percentage units (12.5 means 12.5%).
func ValidateStakeProduct(product *models.TStakeProduct) error {
	if product == nil || product.TenantId <= 0 || strings.TrimSpace(product.ProductNo) == "" ||
		strings.TrimSpace(product.ProductName) == "" || strings.TrimSpace(product.CoinSymbol) == "" ||
		strings.TrimSpace(product.RewardCoinSymbol) == "" {
		return ErrInvalidStakeProduct
	}
	if product.ProductType != int64(staking.ProductType_PRODUCT_TYPE_CURRENT) &&
		product.ProductType != int64(staking.ProductType_PRODUCT_TYPE_FIXED) {
		return ErrInvalidStakeProduct
	}
	if product.InterestMode != int64(staking.InterestMode_INTEREST_MODE_DAILY) &&
		product.InterestMode != int64(staking.InterestMode_INTEREST_MODE_MATURITY) {
		return ErrInvalidStakeProduct
	}
	if product.RewardMode != int64(staking.RewardMode_REWARD_MODE_DAILY) &&
		product.RewardMode != int64(staking.RewardMode_REWARD_MODE_MATURITY) {
		return ErrInvalidStakeProduct
	}
	if product.InterestMode != product.RewardMode {
		return ErrInvalidStakeProduct
	}
	if product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_DISABLE) &&
		product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) &&
		product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_OFF_SHELF) {
		return ErrInvalidStakeProduct
	}
	if product.AllowEarlyRedeem != int64(common.YesNo_YES_NO_YES) &&
		product.AllowEarlyRedeem != int64(common.YesNo_YES_NO_NO) {
		return ErrInvalidStakeProduct
	}
	if !product.Apr.IsPositive() || product.Apr.GreaterThan(decimal.NewFromInt(10000)) ||
		!product.MinAmount.IsPositive() || !product.MaxAmount.IsPositive() ||
		!product.StepAmount.IsPositive() || !product.TotalAmount.IsPositive() ||
		!product.UserLimitAmount.IsPositive() || product.MaxAmount.LessThan(product.MinAmount) ||
		product.UserLimitAmount.LessThan(product.MaxAmount) || product.TotalAmount.LessThan(product.UserLimitAmount) ||
		product.TotalAmount.LessThan(product.StakedAmount) || !product.MinAmount.Mod(product.StepAmount).IsZero() {
		return ErrInvalidStakeProduct
	}
	if product.ProductType == int64(staking.ProductType_PRODUCT_TYPE_FIXED) && product.LockDays <= 0 {
		return ErrInvalidStakeProduct
	}
	if product.ProductType == int64(staking.ProductType_PRODUCT_TYPE_CURRENT) &&
		(product.LockDays != 0 || product.RewardMode != int64(staking.RewardMode_REWARD_MODE_DAILY)) {
		return ErrInvalidStakeProduct
	}
	if product.EarlyRedeemRate.IsNegative() || product.EarlyRedeemRate.GreaterThan(decimal.NewFromInt(100)) {
		return ErrInvalidStakeProduct
	}
	if product.AllowEarlyRedeem == int64(common.YesNo_YES_NO_NO) && !product.EarlyRedeemRate.IsZero() {
		return ErrInvalidStakeProduct
	}
	return nil
}

// ValidateStakeFundingAccounts prevents an enabled product from promising a
// reward that has no configured tenant funding source. Fee revenue is also
// required when early redemption can charge a fee.
func ValidateStakeFundingAccounts(ctx context.Context, svcCtx *svc.ServiceContext, product *models.TStakeProduct) error {
	if product == nil || product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) {
		return nil
	}
	reward, err := svcCtx.AssetAdminClient.GetPlatformAccount(ctx, &asset.GetPlatformAccountReq{
		TenantId: product.TenantId, AccountType: StakingRewardAccountType, Coin: product.RewardCoinSymbol,
	})
	if err != nil || reward.GetBase() == nil || reward.GetBase().GetCode() != 200 || reward.GetData() == nil {
		return ErrInvalidStakeProduct
	}
	balance, err := decimal.NewFromString(reward.GetData().GetAvailableAmount())
	if err != nil || !balance.IsPositive() {
		return ErrInvalidStakeProduct
	}
	if product.AllowEarlyRedeem == int64(common.YesNo_YES_NO_YES) && product.EarlyRedeemRate.IsPositive() {
		fee, feeErr := svcCtx.AssetAdminClient.GetPlatformAccount(ctx, &asset.GetPlatformAccountReq{
			TenantId: product.TenantId, AccountType: FeeRevenueAccountType, Coin: product.CoinSymbol,
		})
		if feeErr != nil || fee.GetBase() == nil || fee.GetBase().GetCode() != 200 || fee.GetData() == nil {
			return ErrInvalidStakeProduct
		}
	}
	return nil
}
