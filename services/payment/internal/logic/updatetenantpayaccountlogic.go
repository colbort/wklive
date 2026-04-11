package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
func (l *UpdateTenantPayAccountLogic) UpdateTenantPayAccount(in *payment.UpdateTenantPayAccountReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdateTenantPayAccount"
	)

	// 査询账户是否存在
	account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if account == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.TenantPayAccountNotFound, l.ctx)),
		}, nil
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
	if in.ExtConfig != "" {
		account.ExtConfig = sql.NullString{String: in.ExtConfig, Valid: true}
	}
	if in.Status != 0 {
		account.Status = int64(in.Status)
	}
	if in.IsDefault != 0 {
		account.IsDefault = in.IsDefault
	}
	if in.Remark != "" {
		account.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	account.UpdateTimes = now

	err = l.svcCtx.TenantPayAccountModel.Update(l.ctx, account)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay account success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
