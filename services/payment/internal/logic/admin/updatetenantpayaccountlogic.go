package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantPayAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayAccountLogic {
	return &UpdateTenantPayAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户支付账号
func (l *UpdateTenantPayAccountLogic) UpdateTenantPayAccount(in *payment.UpdateTenantPayAccountReq) (*payment.CommonResp, error) {
	var (
		errLogic = "UpdateTenantPayAccount"
	)
	if in.Id <= 0 {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}

	// 査询账户是否存在
	account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if errors.Is(err, models.ErrNotFound) || account == nil {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.TenantPayAccountNotFound, i18n.Translate(i18n.TenantPayAccountNotFound, l.ctx)),
		}, nil
	}

	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(l.ctx, account.TenantId, i18n.TenantPayAccountNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate && in.TenantId > 0 {
		account.TenantId = in.TenantId
	}

	now := utils.NowMillis()
	if in.AccountName != "" {
		account.AccountName = in.AccountName
	}
	if in.AppId != "" {
		account.AppId = sql.NullString{String: in.AppId, Valid: true}
	}
	if in.MerchantId != "" {
		account.MerchantId = sql.NullString{String: in.MerchantId, Valid: true}
	}
	if in.MerchantName != "" {
		account.MerchantName = sql.NullString{String: in.MerchantName, Valid: true}
	}
	if in.ApiKeyCipher != "" {
		account.ApiKeyCipher = sql.NullString{String: in.ApiKeyCipher, Valid: true}
	}
	if in.ApiSecretCipher != "" {
		account.ApiSecretCipher = sql.NullString{String: in.ApiSecretCipher, Valid: true}
	}
	if in.PrivateKeyCipher != "" {
		account.PrivateKeyCipher = sql.NullString{String: in.PrivateKeyCipher, Valid: true}
	}
	if in.PublicKey != "" {
		account.PublicKey = sql.NullString{String: in.PublicKey, Valid: true}
	}
	if in.CertCipher != "" {
		account.CertCipher = sql.NullString{String: in.CertCipher, Valid: true}
	}
	if in.CredentialRef != "" {
		account.CredentialRef = in.CredentialRef
	}
	if in.ExtConfig != "" {
		extConfig, valid := nullableJSON(in.ExtConfig)
		if !valid {
			return &payment.CommonResp{
				Base: helper.ErrResp(i18n.InvalidPaymentJSON, i18n.Translate(i18n.InvalidPaymentJSON, l.ctx)),
			}, nil
		}
		account.ExtConfig = extConfig
	}
	if in.Enabled != 0 {
		account.Enabled = int64(in.Enabled)
	}
	if common.YesNo(in.IsDefault) != common.YesNo_YES_NO_UNKNOWN {
		account.IsDefault = int64(in.IsDefault)
	}
	if in.Remark != "" {
		account.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	if !requiredStrings(account.AccountName, account.MerchantId.String, account.MerchantName.String) {
		return paymentErrorResp(l.ctx, i18n.PaymentRequiredParamsMissing), nil
	}
	account.UpdateTimes = now

	err = l.svcCtx.TenantPayAccountModel.Update(l.ctx, account)
	if err != nil {
		if isDuplicateEntry(err) {
			return paymentErrorResp(l.ctx, i18n.TenantPayAccountCodeAlreadyExists), nil
		}
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay account success: %d", in.Id)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}
