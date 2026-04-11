package logic

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
)

type UpdateTenantPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayPlatformLogic {
	return &UpdateTenantPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户开通平台
func (l *UpdateTenantPayPlatformLogic) UpdateTenantPayPlatform(in *payment.UpdateTenantPayPlatformReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "UpdateTenantPayPlatform"
	)

	// 查询租户平台是否存在
	tenantPlatform, err := l.svcCtx.TenantPayPlatformModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if tenantPlatform == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.TenantPlatformNotFound, l.ctx)),
		}, nil
	}

	now := time.Now().UnixMilli()
	if in.Status != 0 {
		tenantPlatform.Status = int64(in.Status)
	}
	if in.OpenStatus != 0 {
		tenantPlatform.OpenStatus = int64(in.OpenStatus)
	}
	if in.Remark != "" {
		tenantPlatform.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	tenantPlatform.UpdateTimes = now

	err = l.svcCtx.TenantPayPlatformModel.Update(l.ctx, tenantPlatform)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Update tenant pay platform success: %d", in.Id)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
