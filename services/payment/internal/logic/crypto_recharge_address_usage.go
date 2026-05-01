package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

const cryptoRechargeAddressHoldSeconds = 180

func cryptoRechargeAddressUsingKey(addressID int64) string {
	return fmt.Sprintf("payment:crypto_recharge_address:using:%d", addressID)
}

func cryptoRechargeAddressUsingValue(tenantID, userID int64) string {
	return fmt.Sprintf("%d:%d", tenantID, userID)
}

func reserveCryptoRechargeAddress(ctx context.Context, svcCtx *svc.ServiceContext, item *models.TCryptoRechargeAddress, tenantID, userID int64) (bool, error) {
	return svcCtx.Redis.SetnxExCtx(ctx, cryptoRechargeAddressUsingKey(item.Id), cryptoRechargeAddressUsingValue(tenantID, userID), cryptoRechargeAddressHoldSeconds)
}

func cryptoRechargeAddressReservedBy(ctx context.Context, svcCtx *svc.ServiceContext, addressID, tenantID, userID int64) (bool, error) {
	key := cryptoRechargeAddressUsingKey(addressID)
	exists, err := svcCtx.Redis.ExistsCtx(ctx, key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	value, err := svcCtx.Redis.GetCtx(ctx, key)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "nil") {
			return false, nil
		}
		return false, err
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return false, nil
	}
	lockedTenantID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, nil
	}
	lockedUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, nil
	}
	return lockedTenantID == tenantID && lockedUserID == userID, nil
}

func refreshCryptoRechargeAddressReservation(ctx context.Context, svcCtx *svc.ServiceContext, addressID int64) {
	if addressID <= 0 {
		return
	}
	_ = svcCtx.Redis.ExpireCtx(ctx, cryptoRechargeAddressUsingKey(addressID), cryptoRechargeAddressHoldSeconds)
}

func releaseCryptoRechargeAddress(ctx context.Context, svcCtx *svc.ServiceContext, addressID int64) {
	if addressID <= 0 {
		return
	}
	_, _ = svcCtx.Redis.DelCtx(ctx, cryptoRechargeAddressUsingKey(addressID))
}
