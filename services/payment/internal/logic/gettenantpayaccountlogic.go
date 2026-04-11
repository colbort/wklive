package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayAccountLogic {
	return &GetTenantPayAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户支付账号详情
func (l *GetTenantPayAccountLogic) GetTenantPayAccount(in *payment.GetTenantPayAccountReq) (*payment.GetTenantPayAccountResp, error) {
	var (
		errLogic = "GetTenantPayAccount"
	)

	account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if account == nil {
		return &payment.GetTenantPayAccountResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.TenantPayAccountNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetTenantPayAccountResp{
		Base: helper.OkResp(),
		Data: &payment.TenantPayAccount{
			Id:                  account.Id,
			TenantId:            account.TenantId,
			TenantPayPlatformId: account.TenantPayPlatformId,
			PlatformId:          account.PlatformId,
			AccountCode:         account.AccountCode,
			AccountName:         account.AccountName,
			AppId:               account.AppId.String,
			MerchantId:          account.MerchantId.String,
			MerchantName:        account.MerchantName.String,
			PublicKey:           account.PublicKey.String,
			ExtConfig:           account.ExtConfig.String,
			Status:              payment.CommonStatus(account.Status),
			IsDefault:           account.IsDefault,
			Remark:              account.Remark.String,
			CreateTimes:         account.CreateTimes,
			UpdateTimes:         account.UpdateTimes,
		},
	}, nil
}
