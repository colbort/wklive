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

type ListTenantPayAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayAccountsLogic {
	return &ListTenantPayAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayAccountsLogic) ListTenantPayAccounts(req *types.ListTenantPayAccountsReq) (resp *types.ListTenantPayAccountsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListTenantPayAccounts(l.ctx, &payment.ListTenantPayAccountsReq{
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:            req.TenantId,
		PlatformId:          req.PlatformId,
		TenantPayPlatformId: req.TenantPayPlatformId,
		Keyword:             req.Keyword,
		Status:              payment.CommonStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.TenantPayAccount, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.TenantPayAccount{
			Id:                  item.Id,
			TenantId:            item.TenantId,
			TenantPayPlatformId: item.TenantPayPlatformId,
			PlatformId:          item.PlatformId,
			AccountCode:         item.AccountCode,
			AccountName:         item.AccountName,
			AppId:               item.AppId,
			MerchantId:          item.MerchantId,
			MerchantName:        item.MerchantName,
			PublicKey:           item.PublicKey,
			ExtConfig:           item.ExtConfig,
			Status:              int64(item.Status),
			IsDefault:           item.IsDefault,
			Remark:              item.Remark,
			CreateTimes:          item.CreateTimes,
			UpdateTimes:          item.UpdateTimes,
		}
	}

	resp = &types.ListTenantPayAccountsResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return resp, nil
}
