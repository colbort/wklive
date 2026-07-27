package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayAccountLogic {
	return &CreateTenantPayAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付账号
func (l *CreateTenantPayAccountLogic) CreateTenantPayAccount(in *payment.CreateTenantPayAccountReq) (*payment.CommonResp, error) {
	var (
		errLogic = "CreateTenantPayAccount"
	)
	tenantID, resp, err := resolveAdminTenantCreateScope(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if in.PlatformId <= 0 || !requiredStrings(in.AccountCode, in.AccountName, in.MerchantId, in.MerchantName) {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	if platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, in.PlatformId); errors.Is(err, models.ErrNotFound) || platform == nil {
		return paymentErrorResp(l.ctx, i18n.PlatformNotFound), nil
	} else if err != nil {
		return nil, err
	}

	now := utils.NowMillis()
	isDefault := int64(in.IsDefault)
	if common.YesNo(in.IsDefault) == common.YesNo_YES_NO_UNKNOWN {
		isDefault = int64(common.YesNo_YES_NO_NO)
	}
	extConfig, valid := helpers.NullableJSON(in.ExtConfig)
	if !valid {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.InvalidPaymentJSON, i18n.Translate(i18n.InvalidPaymentJSON, l.ctx)),
		}, nil
	}
	account := &models.TTenantPayAccount{
		TenantId:         tenantID,
		PlatformId:       in.PlatformId,
		AccountCode:      in.AccountCode,
		AccountName:      in.AccountName,
		AppId:            sql.NullString{String: in.AppId, Valid: true},
		MerchantId:       sql.NullString{String: in.MerchantId, Valid: true},
		MerchantName:     sql.NullString{String: in.MerchantName, Valid: true},
		ApiKeyCipher:     sql.NullString{String: in.ApiKeyCipher, Valid: true},
		ApiSecretCipher:  sql.NullString{String: in.ApiSecretCipher, Valid: true},
		PrivateKeyCipher: sql.NullString{String: in.PrivateKeyCipher, Valid: true},
		PublicKey:        sql.NullString{String: in.PublicKey, Valid: true},
		CertCipher:       sql.NullString{String: in.CertCipher, Valid: true},
		CredentialRef:    in.CredentialRef,
		ExtConfig:        extConfig,
		Enabled:          helpers.EnableToModel(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		IsDefault:        isDefault,
		Remark:           sql.NullString{String: in.Remark, Valid: true},
		CreateTimes:      now,
		UpdateTimes:      now,
	}

	_, err = l.svcCtx.TenantPayAccountModel.Insert(l.ctx, account)
	if err != nil {
		if helpers.IsDuplicateEntry(err) {
			return paymentErrorResp(l.ctx, i18n.TenantPayAccountCodeAlreadyExists), nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Create tenant pay account success: %s", in.AccountCode)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
