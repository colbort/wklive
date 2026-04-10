package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayAccountsLogic {
	return &ListTenantPayAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付账号列表
func (l *ListTenantPayAccountsLogic) ListTenantPayAccounts(in *payment.ListTenantPayAccountsReq) (*payment.ListTenantPayAccountsResp, error) {
	items, total, err := l.svcCtx.TenantPayAccountModel.FindPage(l.ctx, in.TenantId, in.PlatformId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	data := make([]*payment.TenantPayAccount, 0, len(items))
	for _, acc := range items {
		data = append(data, &payment.TenantPayAccount{
			Id:                  acc.Id,
			TenantId:            acc.TenantId,
			TenantPayPlatformId: acc.TenantPayPlatformId,
			PlatformId:          acc.PlatformId,
			AccountCode:         acc.AccountCode,
			AccountName:         acc.AccountName,
			AppId:               acc.AppId.String,
			MerchantId:          acc.MerchantId.String,
			MerchantName:        acc.MerchantName.String,
			PublicKey:           acc.PublicKey.String,
			ExtConfig:           acc.ExtConfig.String,
			Status:              payment.CommonStatus(acc.Status),
			IsDefault:           acc.IsDefault,
			Remark:              acc.Remark.String,
			CreateTimes:         acc.CreateTimes,
			UpdateTimes:         acc.UpdateTimes,
		})
	}

	return &payment.ListTenantPayAccountsResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}
