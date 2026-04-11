package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenTenantPayPlatformLogic {
	return &OpenTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户开通平台
func (l *OpenTenantPayPlatformLogic) OpenTenantPayPlatform(in *payment.OpenTenantPayPlatformReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "OpenTenantPayPlatform"
	)

	now := utils.NowMillis()
	tenantPlatform := &models.TTenantPayPlatform{
		TenantId:    in.TenantId,
		PlatformId:  in.PlatformId,
		Status:      int64(in.Status),
		OpenStatus:  int64(in.OpenStatus),
		Remark:      sql.NullString{String: in.Remark, Valid: true},
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err := l.svcCtx.TenantPayPlatformModel.Insert(l.ctx, tenantPlatform)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Open tenant pay platform success: %d", in.TenantId)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
