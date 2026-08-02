package adminlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	assethelpers "wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
)

const maxBackstopPolicyDuration = int64((366 * 24 * time.Hour) / time.Millisecond)

func platformBackstopPolicyAdmin(ctx context.Context, tenantID int64) (int64, *common.RespBase, error) {
	base, err := assethelpers.AdminTenantWriteScopeResp(ctx, tenantID, i18n.OperationNotAllowed)
	if err != nil || base != nil {
		return 0, base, err
	}
	operatorID, err := utils.GetUserIdFromMd(ctx)
	if err != nil || operatorID <= 0 {
		return 0, nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	return operatorID, nil, nil
}

func parseBackstopPolicyAmount(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, fmt.Errorf("empty backstop policy amount")
	}
	value, err := decimal.NewFromString(raw)
	if err != nil || value.Exponent() < -18 || len(value.Abs().Truncate(0).String()) > 18 {
		return decimal.Zero, fmt.Errorf("invalid backstop policy amount")
	}
	return value, nil
}

func validateBackstopPolicyDraft(
	in *asset.CreatePlatformBackstopPolicyReq,
	now int64,
) (string, string, string, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	if in == nil || in.GetTenantId() <= 0 {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("invalid platform backstop policy")
	}
	coin := strings.ToUpper(strings.TrimSpace(in.GetCoin()))
	requestNo := strings.TrimSpace(in.GetRequestNo())
	reason := strings.TrimSpace(in.GetReason())
	evidenceRef := strings.TrimSpace(in.GetEvidenceRef())
	if coin == "" || len(coin) > 32 || requestNo == "" || len(requestNo) > 96 ||
		reason == "" || len(reason) > 255 || evidenceRef == "" || len(evidenceRef) > 255 ||
		in.GetEffectiveFrom() <= now || in.GetEffectiveUntil() <= in.GetEffectiveFrom() ||
		in.GetEffectiveUntil()-in.GetEffectiveFrom() > maxBackstopPolicyDuration {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("invalid platform backstop policy")
	}
	perRequest, err := parseBackstopPolicyAmount(in.GetPerRequestLimit())
	if err != nil {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	daily, err := parseBackstopPolicyAmount(in.GetDailyLimit())
	if err != nil {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	floor, err := parseBackstopPolicyAmount(in.GetBalanceFloor())
	if err != nil {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	switch in.GetMode() {
	case asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_DISABLED:
		if !perRequest.IsZero() || !daily.IsZero() || !floor.IsZero() {
			return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("disabled backstop policy limits must be zero")
		}
	case asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_PREFUNDED:
		if !perRequest.IsPositive() || !daily.IsPositive() || !floor.IsZero() {
			return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("invalid prefunded backstop policy")
		}
	case asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_CREDIT_FLOOR:
		if !perRequest.IsPositive() || !daily.IsPositive() || !floor.IsNegative() {
			return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("invalid credit-floor backstop policy")
		}
	default:
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("unknown platform backstop policy mode")
	}
	if perRequest.GreaterThan(daily) {
		return "", "", "", decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("per-request limit exceeds daily limit")
	}
	return coin, requestNo, reason, perRequest, daily, floor, nil
}

func sameBackstopPolicyDraft(
	row *models.TAssetBackstopPolicy,
	in *asset.CreatePlatformBackstopPolicyReq,
	coin, reason string,
	perRequest, daily, floor decimal.Decimal,
) bool {
	return row != nil && row.TenantId == in.GetTenantId() && row.Coin == coin &&
		row.Mode == int64(in.GetMode()) && row.PerRequestLimit.Equal(perRequest) &&
		row.DailyLimit.Equal(daily) && row.BalanceFloor.Equal(floor) &&
		row.EffectiveFrom == in.GetEffectiveFrom() && row.EffectiveUntil == in.GetEffectiveUntil() &&
		row.Reason == reason && row.EvidenceRef == strings.TrimSpace(in.GetEvidenceRef())
}

func platformBackstopPolicyProto(row *models.TAssetBackstopPolicy) *asset.PlatformBackstopPolicy {
	if row == nil {
		return nil
	}
	return &asset.PlatformBackstopPolicy{
		Id: row.Id, TenantId: row.TenantId, Coin: row.Coin, RequestNo: row.RequestNo,
		Version: row.Version, Mode: asset.PlatformBackstopMode(row.Mode),
		PerRequestLimit: row.PerRequestLimit.String(), DailyLimit: row.DailyLimit.String(),
		BalanceFloor: row.BalanceFloor.String(), EffectiveFrom: row.EffectiveFrom,
		EffectiveUntil: row.EffectiveUntil, Status: asset.PlatformBackstopPolicyStatus(row.Status),
		Reason: row.Reason, EvidenceRef: row.EvidenceRef, CreatedBy: row.CreatedBy,
		ReviewedBy: row.ReviewedBy, ReviewReason: row.ReviewReason,
		CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes,
	}
}

func backstopPolicyPermissionResp(base *common.RespBase) *asset.PlatformBackstopPolicyResp {
	return &asset.PlatformBackstopPolicyResp{Base: base}
}

func backstopPolicyParamResp(ctx context.Context) *asset.PlatformBackstopPolicyResp {
	return &asset.PlatformBackstopPolicyResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}
