// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantPayAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantPayAccountLogic {
	return &CreateTenantPayAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantPayAccountLogic) CreateTenantPayAccount(req *types.CreateTenantPayAccountReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.CreateTenantPayAccount(l.ctx, &payment.CreateTenantPayAccountReq{
		TenantId:            req.TenantId,
		TenantPayPlatformId: req.TenantPayPlatformId,
		PlatformId:          req.PlatformId,
		AccountCode:         req.AccountCode,
		AccountName:         req.AccountName,
		AppId:               req.AppId,
		MerchantId:          req.MerchantId,
		MerchantName:        req.MerchantName,
		ApiKeyCipher:        req.ApiKeyCipher,
		ApiSecretCipher:     req.ApiSecretCipher,
		PrivateKeyCipher:    req.PrivateKeyCipher,
		PublicKey:           req.PublicKey,
		CertCipher:          req.CertCipher,
		ExtConfig:           req.ExtConfig,
		Status:              payment.CommonStatus(req.Status),
		IsDefault:           req.IsDefault,
		Remark:              req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return resp, nil
}
