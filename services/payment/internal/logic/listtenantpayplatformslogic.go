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

type ListTenantPayPlatformsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayPlatformsLogic {
	return &ListTenantPayPlatformsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户开通平台列表
func (l *ListTenantPayPlatformsLogic) ListTenantPayPlatforms(in *payment.ListTenantPayPlatformsReq) (*payment.ListTenantPayPlatformsResp, error) {
	tenantPlatforms, total, err := l.svcCtx.TenantPayPlatformModel.FindPage(
		l.ctx,
		in.TenantId,
		in.PlatformId,
		int64(in.Status),
		int64(in.OpenStatus),
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(tenantPlatforms)) == in.Page.Limit {
		lastItem := tenantPlatforms[len(tenantPlatforms)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(tenantPlatforms)) == in.Page.Limit

	data := make([]*payment.TenantPayPlatform, 0, len(tenantPlatforms))
	for _, p := range tenantPlatforms {
		data = append(data, &payment.TenantPayPlatform{
			Id:          p.Id,
			TenantId:    p.TenantId,
			PlatformId:  p.PlatformId,
			Status:      payment.CommonStatus(p.Status),
			OpenStatus:  payment.OpenStatus(p.OpenStatus),
			Remark:      p.Remark.String,
			CreateTimes: p.CreateTimes,
			UpdateTimes: p.UpdateTimes,
		})
	}

	return &payment.ListTenantPayPlatformsResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}
