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

type GetTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayPlatformLogic {
	return &GetTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户开通平台详情
func (l *GetTenantPayPlatformLogic) GetTenantPayPlatform(in *payment.GetTenantPayPlatformReq) (*payment.GetTenantPayPlatformResp, error) {
	var (
		errLogic = "GetTenantPayPlatform"
	)

	tenantPlatform, err := l.svcCtx.TenantPayPlatformModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if tenantPlatform == nil {
		return &payment.GetTenantPayPlatformResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.TenantPlatformNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetTenantPayPlatformResp{
		Base: helper.OkResp(),
		Data: toTenantPayPlatformProto(tenantPlatform),
	}, nil
}
