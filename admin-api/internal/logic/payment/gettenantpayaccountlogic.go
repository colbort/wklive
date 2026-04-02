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

type GetTenantPayAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayAccountLogic {
	return &GetTenantPayAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayAccountLogic) GetTenantPayAccount(req *types.GetTenantPayAccountReq) (resp *types.GetTenantPayAccountResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetTenantPayAccount(l.ctx, &payment.GetTenantPayAccountReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetTenantPayAccountResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.TenantPayAccount{
			Id:                  result.Data.Id,
			TenantId:            result.Data.TenantId,
			TenantPayPlatformId: result.Data.TenantPayPlatformId,
			PlatformId:          result.Data.PlatformId,
			AccountCode:         result.Data.AccountCode,
			AccountName:         result.Data.AccountName,
			AppId:               result.Data.AppId,
			MerchantId:          result.Data.MerchantId,
			MerchantName:        result.Data.MerchantName,
			PublicKey:           result.Data.PublicKey,
			ExtConfig:           result.Data.ExtConfig,
			Status:              int64(result.Data.Status),
			IsDefault:           result.Data.IsDefault,
			Remark:              result.Data.Remark,
			CreateTime:          result.Data.CreateTime,
			UpdateTime:          result.Data.UpdateTime,
		},
	}
	return resp, nil
}
