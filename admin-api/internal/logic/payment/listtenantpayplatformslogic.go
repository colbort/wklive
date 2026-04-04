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

type ListTenantPayPlatformsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayPlatformsLogic {
	return &ListTenantPayPlatformsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantPayPlatformsLogic) ListTenantPayPlatforms(req *types.ListTenantPayPlatformsReq) (resp *types.ListTenantPayPlatformsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListTenantPayPlatforms(l.ctx, &payment.ListTenantPayPlatformsReq{
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:   req.TenantId,
		PlatformId: req.PlatformId,
		Status:     payment.CommonStatus(req.Status),
		OpenStatus: payment.OpenStatus(req.OpenStatus),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.TenantPayPlatform, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.TenantPayPlatform{
			Id:         item.Id,
			TenantId:   item.TenantId,
			PlatformId: item.PlatformId,
			Status:     int64(item.Status),
			OpenStatus: int64(item.OpenStatus),
			Remark:     item.Remark,
			CreateTimes: item.CreateTimes,
			UpdateTimes: item.UpdateTimes,
		}
	}

	resp = &types.ListTenantPayPlatformsResp{
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
