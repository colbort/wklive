package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDetailLogic {
	return &SysTenantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据 code 获取租户
func (l *SysTenantDetailLogic) SysTenantDetail(in *system.SysTenantDetailReq) (*system.SysTenantDetailResp, error) {
	if in.TenantId == nil && in.TenantCode == nil {
		return &system.SysTenantDetailResp{
			Base: helper.FailWithCode(i18n.ParamError),
		}, nil
	}
	var result *models.SysTenant
	var err error
	if in.TenantId != nil {
		result, err = l.svcCtx.TenantMode.FindOne(l.ctx, *in.TenantId)
	} else if in.TenantCode != nil {
		result, err = l.svcCtx.TenantMode.FindByTenantCode(l.ctx, *in.TenantCode)
	}
	if err != nil {
		return nil, err
	}
	return &system.SysTenantDetailResp{
		Base: helper.OkResp(),
		Data: &system.SysTenantItem{
			Id:           result.Id,
			TenantCode:   result.TenantCode,
			TenantName:   result.TenantName,
			Status:       commonStatusToProto(result.Status),
			ExpireTime:   result.ExpireTime,
			ContactName:  result.ContactName.String,
			ContactPhone: result.ContactPhone.String,
			Remark:       result.Remark.String,
			CreateTimes:  result.CreateTimes,
			UpdateTimes:  result.UpdateTimes,
		},
	}, nil
}
