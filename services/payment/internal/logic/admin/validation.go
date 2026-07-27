package adminlogic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/shopspring/decimal"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

func paymentErrorResp(ctx context.Context, code int32) *payment.CommonResp {
	return &payment.CommonResp{Base: helper.ErrResp(code, i18n.Translate(code, ctx))}
}

func requiredStrings(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func validNonNegativeRange(minValue, maxValue int64) bool {
	return minValue >= 0 && maxValue >= 0 && (maxValue == 0 || minValue <= maxValue)
}

func parseNonNegativeAmount(value string) (decimal.Decimal, error) {
	amount, err := conv.ParseDecimalField(value)
	if err != nil || amount.IsNegative() {
		return decimal.Zero, errors.New("invalid non-negative payment amount")
	}
	return amount, nil
}

func validDecimalRange(minValue, maxValue decimal.Decimal) bool {
	return !minValue.IsNegative() && !maxValue.IsNegative() &&
		(maxValue.IsZero() || !minValue.GreaterThan(maxValue))
}

func validJSONArray(value string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	var items []any
	return json.Unmarshal([]byte(value), &items) == nil
}

func resolveAdminTenantCreateScope(ctx context.Context, requestedTenantID int64) (int64, *payment.CommonResp, error) {
	userType, err := utils.GetUserTypeFromMd(ctx)
	if err != nil {
		return 0, nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	switch userType {
	case utils.SysUserTypeSystemAdmin:
		if requestedTenantID <= 0 {
			return 0, paymentErrorResp(ctx, i18n.PaymentRequiredParamsMissing), nil
		}
		return requestedTenantID, nil, nil
	case utils.SysUserTypeTenantOwner, utils.SysUserTypeTenantAdmin:
		tenantID, err := utils.GetTenantIdFromMd(ctx)
		if err != nil {
			return 0, nil, i18n.StatusError(ctx, i18n.UserNotFound)
		}
		if requestedTenantID > 0 && requestedTenantID != tenantID {
			return 0, paymentErrorResp(ctx, i18n.PermissionDenied), nil
		}
		return tenantID, nil, nil
	default:
		return 0, paymentErrorResp(ctx, i18n.PermissionDenied), nil
	}
}

func validateProductPlatform(ctx context.Context, svcCtx *svc.ServiceContext, productID, platformID int64) (*payment.CommonResp, error) {
	product, err := svcCtx.PayProductModel.FindOne(ctx, productID)
	if errors.Is(err, models.ErrNotFound) || product == nil {
		return paymentErrorResp(ctx, i18n.ProductNotFound), nil
	}
	if err != nil {
		return nil, err
	}
	if product.PlatformId != platformID {
		return paymentErrorResp(ctx, i18n.PaymentRelationMismatch), nil
	}
	return nil, nil
}

func validateAccountPlatform(ctx context.Context, svcCtx *svc.ServiceContext, accountID, tenantID, platformID int64) (*payment.CommonResp, error) {
	account, err := svcCtx.TenantPayAccountModel.FindOne(ctx, accountID)
	if errors.Is(err, models.ErrNotFound) || account == nil {
		return paymentErrorResp(ctx, i18n.TenantPayAccountNotFound), nil
	}
	if err != nil {
		return nil, err
	}
	if account.TenantId != tenantID || account.PlatformId != platformID {
		return paymentErrorResp(ctx, i18n.PaymentRelationMismatch), nil
	}
	return nil, nil
}

func validateChannelTenant(ctx context.Context, svcCtx *svc.ServiceContext, channelID, tenantID int64) (*payment.CommonResp, error) {
	channel, err := svcCtx.TenantPayChannelModel.FindOne(ctx, channelID)
	if errors.Is(err, models.ErrNotFound) || channel == nil {
		return paymentErrorResp(ctx, i18n.PaymentChannelNotFound), nil
	}
	if err != nil {
		return nil, err
	}
	if channel.TenantId != tenantID {
		return paymentErrorResp(ctx, i18n.PaymentRelationMismatch), nil
	}
	return nil, nil
}
