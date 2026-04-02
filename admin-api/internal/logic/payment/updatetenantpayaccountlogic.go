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

type UpdateTenantPayAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayAccountLogic {
	return &UpdateTenantPayAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayAccountLogic) UpdateTenantPayAccount(req *types.UpdateTenantPayAccountReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdateTenantPayAccount(l.ctx, &payment.UpdateTenantPayAccountReq{
		Id:               req.Id,
		TenantId:         req.TenantId,
		AccountName:      req.AccountName,
		AppId:            req.AppId,
		MerchantId:       req.MerchantId,
		MerchantName:     req.MerchantName,
		ApiKeyCipher:     req.ApiKeyCipher,
		ApiSecretCipher:  req.ApiSecretCipher,
		PrivateKeyCipher: req.PrivateKeyCipher,
		PublicKey:        req.PublicKey,
		CertCipher:       req.CertCipher,
		ExtConfig:        req.ExtConfig,
		Status:           payment.CommonStatus(req.Status),
		IsDefault:        req.IsDefault,
		Remark:           req.Remark,
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
